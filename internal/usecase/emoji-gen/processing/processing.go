package processing

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func (m *Module) RoundDimensions(width, height int) (int, int) {
	newWidth := (width / 100) * 100
	newHeight := int(float64(height) * (float64(newWidth) / float64(width)))
	if newHeight%2 != 0 {
		newHeight--
	}
	return newWidth, newHeight
}

func (m *Module) RoundUpTo100(num int) int {
	return ((num + 99) / 100) * 100
}

func (m *Module) DimensionToNewWidth(width, height, newWidth int) (int, int) {
	newHeight := int(float64(height) * (float64(newWidth) / float64(width)))
	return newWidth, newHeight
}

func (m *Module) getVideoDimensions(inputVideo string) (width, height int, err error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-count_packets",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		inputVideo)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	_, err = fmt.Sscanf(string(output), "%d,%d", &width, &height)
	if err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

type tile struct {
	OutputFile string
	FFmpegArgs []string
	Position   int
}

type processResult struct {
	filename string
	position int
	err      error
}

func (m *Module) ProcessVideo(ctx context.Context, args *entity.EmojiCommand) ([]string, error) {
	// Проверяем отмену
	select {
	case <-ctx.Done():
		return nil, nil
	default:
	}

	width, height, err := m.getVideoDimensions(args.DownloadedFile)
	if err != nil {
		return nil, err
	}

	if args.QualityValue == 0 {
		if width >= 100 && height >= 100 {
			width, height = m.RoundDimensions(width, height)
		} else {
			if width < 100 {
				width = 100
			}
			if height < 100 {
				height = 100
			}
		}

		if args.Width != 0 {
			width, height = m.DimensionToNewWidth(width, height, args.Width*100)
		}
		var i int
		for i = width; i >= 100; i = i / 100 {
		}
		args.Width = i

		args.DownloadedFile, err = m.resizeVideo(args, width, height)
		if err != nil {
			return nil, err
		}
	}

	// Проверяем отмену после ресайза
	select {
	case <-ctx.Done():
		return nil, nil
	default:
	}

	lastRowHeight := height % 100
	if lastRowHeight < 10 && lastRowHeight > 0 {
		args.DownloadedFile, err = m.cropVideoHeight(args.DownloadedFile, args.WorkingDir, height)
		if err != nil {
			return nil, err
		}
		_, height, err = m.getVideoDimensions(args.DownloadedFile)
		if err != nil {
			return nil, err
		}
	}

	tileWidth := 100
	tileHeight := 100
	tilesX := width / tileWidth
	tilesY := height / tileHeight
	lastRowHeight = height % 100 // Пересчитываем lastRowHeight после возможного кропа
	if lastRowHeight > 0 {
		tilesY++
	}

	baseFFmpegArgs := []string{
		"-y",
		"-i", args.DownloadedFile,
		"-c:v", "libvpx-vp9",
		"-profile:v", "0",
		"-pix_fmt", "yuva420p",
		"-crf", "24",
		"-b:v", fmt.Sprintf("%d", args.QualityValue),
		"-b:a", "256k",
		"-t", "3.0",
		"-r", "10",
		"-auto-alt-ref", "1",
		"-metadata:s:v:0", "alpha_mode=1",
		"-an",
	}

	jobs := make(chan tile, tilesX*tilesY)
	results := make(chan processResult, tilesX*tilesY)

	numWorkers := 4
	var wg sync.WaitGroup

	// Запускаем воркеров с контекстом
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go worker(ctx, jobs, results, &wg)
	}

	// Отдельная горутина для отправки заданий
	go func() {
		position := 0
		for j := 0; j < tilesY; j++ {
			for i := 0; i < tilesX; i++ {
				// Проверяем отмену перед отправкой каждого задания
				select {
				case <-ctx.Done():
					close(jobs)
					return
				default:
				}

				outputFile := filepath.Join(args.WorkingDir, fmt.Sprintf("emoji_%d_%d.webm", j, i))
				var vfArgs []string

				currentTileHeight := tileHeight
				if j == tilesY-1 && lastRowHeight > 0 {
					currentTileHeight = lastRowHeight
					padColor := "#04F404@0.1"
					if args.BackgroundColor != "" {
						padColor = args.BackgroundColor
					} else if args.BackgroundBlend != "" {
						args.BackgroundColor = padColor
					}
					vfArgs = []string{
						fmt.Sprintf("crop=%d:%d:%d:%d", tileWidth, currentTileHeight, i*tileWidth, j*tileHeight),
						fmt.Sprintf("scale=100:%d", currentTileHeight),
						fmt.Sprintf("pad=100:100:%d:0:color=%s", 100-currentTileHeight, padColor),
					}
				} else {
					vfArgs = []string{
						fmt.Sprintf("crop=%d:%d:%d:%d", tileWidth, currentTileHeight, i*tileWidth, j*tileHeight),
					}
				}

				if args.BackgroundColor != "" {
					vfArgs = append(vfArgs, fmt.Sprintf("colorkey=%s:similarity=%s:blend=%s", args.BackgroundColor, args.BackgroundSim, args.BackgroundBlend))
				}
				vfArgs = append(vfArgs, "setsar=1:1")
				ffmpegArgs := make([]string, len(baseFFmpegArgs))
				copy(ffmpegArgs, baseFFmpegArgs)
				ffmpegArgs = append(ffmpegArgs, "-vf", strings.Join(vfArgs, ","), outputFile)

				jobs <- tile{
					OutputFile: outputFile,
					FFmpegArgs: ffmpegArgs,
					Position:   position,
				}
				position++
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	resultSlice := make([]string, tilesX*tilesY)
	var errors []error

	// Обрабатываем результаты с учетом отмены
	for {
		select {
		case <-ctx.Done():
			return nil, nil
		case result, ok := <-results:
			if !ok {
				// Канал закрыт, все результаты получены
				if len(errors) > 0 {
					return nil, fmt.Errorf("encountered %d errors during processing", len(errors))
				}

				// Убираем пустые элементы из результата
				var finalResults []string
				for _, res := range resultSlice {
					if res != "" {
						finalResults = append(finalResults, res)
					}
				}
				return finalResults, nil
			}
			if result.err != nil {
				errors = append(errors, result.err)
				log.Printf("Error during processing: %v", result.err)
				continue
			}
			resultSlice[result.position] = result.filename
		}
	}
}

func worker(ctx context.Context, jobs <-chan tile, results chan<- processResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			// Создаем команду с возможностью отмены
			cmd := exec.CommandContext(ctx, "ffmpeg", job.FFmpegArgs...)
			err := cmd.Run()

			// Проверяем отмену перед отправкой результата
			select {
			case <-ctx.Done():
				return
			default:
				results <- processResult{
					filename: job.OutputFile,
					position: job.Position,
					err:      err,
				}
			}
		}
	}
}

func (m *Module) cropVideoHeight(inputFile string, workingDir string, height int) (string, error) {
	outputFile := filepath.Join(workingDir, "cropped.webm")

	// Вычисляем новую высоту (округляем вниз до ближайшей сотни)
	newHeight := (height / 100) * 100

	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-c:v", "libvpx-vp9",
		"-vf", fmt.Sprintf("crop=iw:%.0f:0:0", float64(newHeight)),
		"-y",
		outputFile)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ошибка при обрезке видео: %w", err)
	}

	return outputFile, nil
}

func (m *Module) resizeVideo(args *entity.EmojiCommand, toWidth, toHeight int) (string, error) {
	outputFile := filepath.Join(args.WorkingDir, "resized.webm")

	cmd := exec.Command("ffmpeg",
		"-i", args.DownloadedFile,
		"-c:v", "libvpx-vp9",
		"-vf", fmt.Sprintf("scale=%d:%d", toWidth, toHeight),
		"-y",
		outputFile)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ошибка при изменении размера файла: %w", err)
	}

	return outputFile, nil
}
