
# ELYSIUM MANAGER

## Токены

Arima DJ  7447550770:AAHaO6tRmqNtb53fD9cIXPJVjYi1mHN3i_0 
Demethra 7445477091:AAGOqZ_0_5vTkhHRNfK2iHWgk4ejM8UkL_8
Demethra Test 7486051673:AAGXMsNZ3ia99ljU48IErrA5PH4ZV-VncFo

## HOWTO
### Миграции

##### Накатить новую миграцию

1. Добавить новый файл миграции\
   `migrate create -ext sql -dir ./migrations -seq {name_of_migration_file} `
2. Описать миграцию внутри файла
3. Накатить миграцию\
   `migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations up 1`

Если возника ошибка, необходимо исправить ошибку в файле миграции и зафорсить предыдущую версию миграции и накатить новую:

Попробуйте откатить текущую миграцию:\
`migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations down 1`\

Вам вернется ошибка:\
`Dirty database version 2. Fix and force version.`\
*2* - в этом случае текущая версия базы.\

Необходимо откатить на прошлую. Следует выполнить сначала команду `force` с  текущей версией (в данном случае 2):\
`migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations force 2`


Далее откатить на 1 миграцию назад (не на версию 1 а на одну версию вниз):\
`migrate -database 'postgres://login:password@addr:port/db_name?sslmode=disable' -path ./migrations down 1`

## Скрипт для Tapermonkey
    
```javascript
// ==UserScript==
// @name         Get SoundCloud Track Info 2
// @namespace    http://tampermonkey.net/
// @version      0.1
// @description  Extracts track title and artist name from SoundCloud every second and sends it to a server
// @author       You
// @match        *://soundcloud.com/*
// @grant        none
// ==/UserScript==

(function() {
    'use strict';

    const apiKey = '1234515151';
    // Функция для извлечения названия трека и имени артиста
    function getTrackInfo() {
        const trackTitleElement = document.querySelector('.playbackSoundBadge__titleLink span');
        const artistNameElement = document.querySelector('.playbackSoundBadge__lightLink');
        const trackLinkElement = document.querySelector('.playbackSoundBadge__titleLink');
        const durationElement = document.querySelector('.playbackTimeline__duration span[aria-hidden="true"]');
//         const artworkUrlElement = document.evaluate("//meta[@property='og:image']/@content", document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue;/
//             const artworkUrlElement = document.querySelector('meta[property="og:image"]');
         const artworkSpanElement = document.querySelector('.playbackSoundBadge__avatar span');

                const releaseDateElement = document.querySelector('.soundTitle__uploadTime span[datetime]');
        const tagElements = document.querySelectorAll('.sc-tag');

        const trackTitle = trackTitleElement ? trackTitleElement.textContent : 'Unknown';
        const artistName = artistNameElement ? artistNameElement.textContent : 'Unknown';
        const trackLink = trackLinkElement ? trackLinkElement.href : 'Unknown';
        const duration = durationElement ? durationElement.textContent.trim() : '05:31';
       // const artworkUrl = artworkUrlElement ? artworkUrlElement.getAttribute('content') : '';
         const releaseDate = releaseDateElement ? releaseDateElement.getAttribute('datetime') : '';
        const tags = Array.from(tagElements).map(tag => tag.textContent.trim());

        let artworkUrl = '';
        if (artworkSpanElement) {
            const backgroundImage = artworkSpanElement.style.backgroundImage;
            artworkUrl = backgroundImage.slice(5, -2); // Убираем 'url(' и ')'
        }

        console.log(artworkUrl)
        return { trackTitle, artistName, trackLink, duration, artworkUrl, releaseDate, tags};
    }

    // Отправка данных на сервер
    function sendTrackInfoToServer(trackInfo) {
        fetch('https://elysiumfm.ru/api/tampermonkey/submit', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'x-api-key': apiKey,
            },
            body: JSON.stringify(trackInfo),
        });
    }

    // Извлечение данных и отправка на сервер каждые 1000 мс (1 секунда)
    setInterval(function() {
        const trackInfo = getTrackInfo();
        if (trackInfo.trackLink === 'unknown') {return};
        sendTrackInfoToServer(trackInfo);
    }, 10000);

})();
```

Version 2

```javascript
// ==UserScript==
// @name         Get Music Track Info 2
// @namespace    http://tampermonkey.net/
// @version      0.8
// @description  Extracts track title and artist name from SoundCloud, Spotify, and YouTube Music every 10 seconds
// @author       You
// @match        *://soundcloud.com/*
// @match        *://open.spotify.com/*
// @match        *://music.youtube.com/*
// @grant        none
// ==/UserScript==

(function() {
   'use strict';

   let spotifyAccessToken = '';
   let spotifyTokenExpiration = 0;
   const apiKey = '1234515151';

   // Функция для извлечения информации о треке в зависимости от сайта
   async function getTrackInfo() {
      const url = window.location.href;
      return await extractTrackInfo(url);
   }

   async function extractTrackInfo(url) {
      if (url.includes('soundcloud.com')) {
         const trackTitleElement = document.querySelector('.playbackSoundBadge__titleLink span');
         const artistNameElement = document.querySelector('.playbackSoundBadge__lightLink');
         const trackLinkElement = document.querySelector('.playbackSoundBadge__titleLink');
         const durationElement = document.querySelector('.playbackTimeline__duration span[aria-hidden="true"]');
         const artworkSpanElement = document.querySelector('.playbackSoundBadge__avatar span');

         const trackTitle = trackTitleElement ? trackTitleElement.textContent : 'Unknown';
         const artistName = artistNameElement ? artistNameElement.textContent : 'Unknown';
         const trackLink = trackLinkElement ? trackLinkElement.href : 'Unknown';
         const duration = durationElement ? durationElement.textContent.trim() : 'Unknown';

         let artworkUrl = '';
         if (artworkSpanElement) {
            const backgroundImage = artworkSpanElement.style.backgroundImage;
            artworkUrl = backgroundImage.slice(5, -2); // Убираем 'url(' и ')'
         }

         return { trackTitle, artistName, trackLink, duration, artworkUrl };
      } else if (url.includes('open.spotify.com')) {
         const trackTitleElement = document.querySelector('[data-testid="context-item-link"]');
         const artistNameElement = document.querySelector('[data-testid="context-item-info-artist"]');
         const durationElement = document.querySelector('[data-testid="playback-duration"]');

         const trackTitle = trackTitleElement ? trackTitleElement.textContent : 'Unknown';
         const artistName = artistNameElement ? artistNameElement.textContent : 'Unknown';
         const duration = durationElement ? durationElement.textContent.trim() : 'Unknown';

         const spotifyInfo = await searchSpotifyTrack(artistName, trackTitle);

         return {
            trackTitle,
            artistName,
            trackLink: spotifyInfo.trackLink,
            duration,
            artworkUrl: spotifyInfo.artworkUrl
         };
      } else if (url.includes('music.youtube.com')) {
         const playerBar = document.querySelector('ytmusic-player-bar');

         if (playerBar) {
            const trackTitleElement = playerBar.querySelector('.title');
            const artistNameElement = playerBar.querySelector('.byline > a');
            const artworkElement = playerBar.querySelector('.image');
            const durationElement = playerBar.querySelector('.time-info');
            const trackLinkElement = document.querySelector('.ytp-title-link');

            const trackTitle = trackTitleElement ? trackTitleElement.textContent.trim() : 'Unknown';
            const artistName = artistNameElement ? artistNameElement.textContent.trim() : 'Unknown';
            const artworkUrl = artworkElement ? artworkElement.src : '';
            const duration = durationElement ? durationElement.textContent.trim().split(' / ').at(-1) : 'Unknown';

            let trackLink = '';
            if (trackLinkElement) {
               const urlObj = new URL(trackLinkElement.href);
               const vParam = urlObj.searchParams.get('v');
               urlObj.search = '';
               if (vParam) {
                  urlObj.searchParams.set('v', vParam);
               }
               trackLink = urlObj.toString();
            }

            return { trackTitle, artistName, trackLink, duration, artworkUrl };
         }
      }

      return null;
   }

   // Отправка данных на сервер
   function sendTrackInfoToServer(trackInfo) {
      fetch('https://elysiumfm.ru/api/tampermonkey/submit', {
         method: 'POST',
         headers: {
            'Content-Type': 'application/json',
            'x-api-key': apiKey,
         },
         body: JSON.stringify(trackInfo),
      });
      
       console.log(
              trackInfo.trackTitle,
              trackInfo.artistName,
              trackInfo.trackLink,
              trackInfo.duration,
              trackInfo.artworkUrl
      );
      // Здесь вы можете добавить код для отправки данных на ваш сервер
   }

   const searchSpotifyTrack = async (artistName, trackName) => {
      if (!spotifyAccessToken || Date.now() >= spotifyTokenExpiration) {
         if (!extractSpotifyAccessToken()) {
            console.error('Spotify access token not available');
            return { trackLink: 'Spotify access token not available', artworkUrl: '' };
         }
      }

      const query = encodeURIComponent(`artist:${artistName} track:${trackName}`);

      try {
         const response = await fetch(
                 `https://api.spotify.com/v1/search?q=${query}&type=track&limit=1`,
                 {
                    headers: {
                       Authorization: `Bearer ${spotifyAccessToken}`
                    }
                 }
         );

         if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
         }

         const data = await response.json();

         if (data.tracks.items.length > 0) {
            const track = data.tracks.items[0];
            return {
               trackLink: track.external_urls.spotify,
               artworkUrl: track.album.images[0]?.url || ''
            };
         } else {
            return { trackLink: 'Трек не найден', artworkUrl: '' };
         }
      } catch (error) {
         console.error('Ошибка при поиске трека:', error);
         return { trackLink: 'Произошла ошибка при поиске трека', artworkUrl: '' };
      }
   };

   // Функция для извлечения access token из Spotify
   function extractSpotifyAccessToken() {
      const sessionScript = document.getElementById('session');
      if (sessionScript) {
         try {
            const sessionData = JSON.parse(sessionScript.textContent);
            spotifyAccessToken = sessionData.accessToken;
            spotifyTokenExpiration = sessionData.accessTokenExpirationTimestampMs;
            return true;
         } catch (error) {
            console.error('Error parsing Spotify session data:', error);
         }
      }
      console.error('Failed to extract Spotify access token');
      return false;
   }

   // Извлечение данных и отправка на сервер каждые 10 секунд
   setInterval(async function() {
      if (window.location.href.includes('open.spotify.com')) {
         extractSpotifyAccessToken();
      }
      const trackInfo = await getTrackInfo();
      if (trackInfo) {
         sendTrackInfoToServer(trackInfo);
      }
   }, 15000);

   // Инициализация: попытка извлечь токен при загрузке скрипта на Spotify
   if (window.location.href.includes('open.spotify.com')) {
      extractSpotifyAccessToken();
   }
})();

```

## Макет (Layout)

Функциональность макета в этом проекте позволяет пользователям создавать и управлять настраиваемыми макетами. Вот обзор того, как это работает:

### Структура

- `UserLayout`: Представляет макет пользователя, включая фон, элементы макета, создателя и редакторов.
- `LayoutElement`: Представляет отдельный элемент в макете, с такими свойствами, как позиция, тип и видимость.

### Разрешения

- Создатель: Пользователь, создавший макет. Имеет полные права на редактирование.
- Редакторы: Пользователи, которым предоставлено разрешение на редактирование макета.
- Просмотрщики: Все остальные пользователи, которые могут просматривать макет, но не могут его редактировать.

### Публичные и приватные элементы

- Каждый `LayoutElement` имеет булево поле `Public`.
- Если `Public` равно true, элемент виден всем пользователям.
- Если `Public` равно false, элемент виден только создателю и редакторам.

### Ключевые операции

1. **GetUserLayout**: 
   - Получает макет пользователя.
   - Фильтрует приватные элементы, если запрашивающий не имеет прав на редактирование.

2. **UpdateLayoutFull**: 
   - Обновляет весь макет.
   - Требует прав на редактирование.
   - Логирует изменение.

3. **AddEditor**: 
   - Добавляет нового редактора в макет.
   - Только создатель может выполнить это действие.
   - Логирует изменение.

4. **RemoveEditor**: 
   - Удаляет редактора из макета.
   - Только создатель может выполнить это действие.
   - Логирует изменение.

5. **IsEditor**: 
   - Проверяет, имеет ли пользователь права на редактирование конкретного макета.

### Логирование

Все значительные изменения макетов (обновления, добавление/удаление редакторов) логируются с помощью метода `LogLayoutChange`. Это помогает отслеживать историю изменений, внесенных в макеты.

### Лучшие практики

- Всегда проверяйте разрешения перед тем, как разрешить редактирование макета.
- Используйте поле `Public` для контроля видимости чувствительных элементов макета.
- Логируйте все значительные изменения для поддержания аудиторского следа.

Для более подробной информации о каждой операции обратитесь к документации методов в пакете `layout`.
