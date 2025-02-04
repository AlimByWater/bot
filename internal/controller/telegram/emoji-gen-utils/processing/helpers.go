package processing

import (
	"elysium/internal/entity"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	outputDirTemplate      = "/tmp/%s"
	defaultBackgroundSim   = "0.1"
	defaultBackgroundBlend = "0.1"
)

func (m *Module) ExtractCommandArgs(msgText, msgCaption string) string {
	var args string
	if strings.HasPrefix(msgText, "/emoji") {
		args = strings.TrimPrefix(msgText, "/emoji")
	} else if strings.HasPrefix(msgCaption, "/emoji ") {
		args = strings.TrimPrefix(msgCaption, "/emoji ")
	}
	return strings.TrimSpace(args)
}

func (m *Module) SetupEmojiCommand(args *entity.EmojiCommand, user entity.User) *entity.EmojiCommand {
	// Set default values
	if args.Width == 0 {
		args.Width = entity.DefaultWidth
	}
	if args.BackgroundSim == "" {
		args.BackgroundSim = defaultBackgroundSim
	}
	if args.BackgroundBlend == "" {
		args.BackgroundBlend = defaultBackgroundBlend
	}

	if args.SetTitle == "" {
		args.SetTitle = strings.TrimSpace(entity.PackTitleTempl)
	} else {
		if args.Permissions.Vip {
			args.SetTitle = strings.TrimSpace(args.SetTitle)
		} else {
			if len(args.SetTitle) > entity.TelegramPackLinkAndNameLength-len(entity.PackTitleTempl) {
				args.SetTitle = args.SetTitle[:entity.TelegramPackLinkAndNameLength-len(entity.PackTitleTempl)]
			}
			args.SetTitle = fmt.Sprintf(`%s
%s`, args.SetTitle, entity.PackTitleTempl)
		}

	}

	// Setup working directory and user info
	postfix := fmt.Sprintf("%d_%d", user.TelegramID, time.Now().Unix())
	args.WorkingDir = fmt.Sprintf(outputDirTemplate, postfix)

	args.TelegramUserID = user.TelegramID
	args.UserName = user.TelegramUsername

	return args
}

func (m *Module) ParseArgs(arg string) (*entity.EmojiCommand, error) {
	var emojiArgs entity.EmojiCommand
	emojiArgs.RawInitCommand = "/emoji " + arg

	if arg == "" {
		emojiArgs.SetDefault()
		return &emojiArgs, nil
	}

	var args []string
	currentArg := ""
	inBrackets := false

	arg = strings.ReplaceAll(arg, "\n", " ")

	// Проходим по строке посимвольно для корректной обработки значений в скобках
	for i := 0; i < len(arg); i++ {
		switch arg[i] {
		case '[':
			inBrackets = true
		case ']':
			inBrackets = false
		case ' ':
			if !inBrackets {
				if currentArg != "" {
					args = append(args, currentArg)
					currentArg = ""
				}
			} else {
				currentArg += string(arg[i])
			}
		default:
			currentArg += string(arg[i])
		}
	}
	if currentArg != "" {
		args = append(args, currentArg)
	}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue // Пропускаем несуществующий аргумент
		}

		key := strings.ToLower(parts[0])
		value := parts[1]

		// Определяем стандартный ключ из алиаса
		standardKey, exists := entity.ArgAlias[key]
		if !exists {
			continue // Пропускаем несуществующий аргумент
		}

		// Обрабатываем аргумент в зависимости от стандартного ключа
		switch standardKey {
		case "width":
			width, err := strconv.Atoi(value)
			if err != nil {
				continue
			}
			emojiArgs.Width = width
		case "name":
			emojiArgs.SetTitle = strings.TrimSpace(value)
		case "background":
			emojiArgs.BackgroundColor = m.ColorToHex(value)
		case "background_blend":
			value = strings.ReplaceAll(value, ",", ".")
			emojiArgs.BackgroundBlend = value
		case "background_sim":
			value = strings.ReplaceAll(value, ",", ".")
			emojiArgs.BackgroundSim = value
		case "link":
			emojiArgs.PackLink = value
		case "iphone":
			if value != "true" && value != "false" {
				value = "true"
			}
			emojiArgs.Iphone = value == "true"
		}
	}

	if (emojiArgs.BackgroundSim != "" || emojiArgs.BackgroundBlend != "") && emojiArgs.BackgroundColor == "" {
		return &emojiArgs, entity.ErrInvalidBackgroundArgumentsUse
	}

	return &emojiArgs, nil
}

func (m *Module) ColorToHex(colorName string) string {
	if colorName == "" {
		return ""
	}

	if hex, exists := entity.ColorMap[strings.ToLower(colorName)]; exists {
		return hex
	}

	// Если это уже hex формат или неизвестный цвет, возвращаем как есть
	if strings.HasPrefix(colorName, "0x") {
		return colorName
	}

	if strings.HasPrefix(colorName, "0X") {
		colorName = strings.Replace(colorName, "0X", "0x", 1)
		return colorName
	}

	if strings.HasPrefix(colorName, "#") {
		colorName = strings.Replace(colorName, "#", "0x", 1)
		return colorName
	}

	fmt.Println(colorName)

	if !strings.HasPrefix(colorName, "0x") && !strings.HasPrefix(colorName, "#") {
		colorName = "0x" + colorName
		return colorName
	}

	return "0x000000" // возвращаем черный по умолчанию
}
