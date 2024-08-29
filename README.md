
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
// @name         Get SoundCloud Track Info
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
             const artworkUrlElement = document.querySelector('meta[property="og:image"]');
                const releaseDateElement = document.querySelector('.soundTitle__uploadTime span[datetime]');
        const tagElements = document.querySelectorAll('.sc-tag');

        const trackTitle = trackTitleElement ? trackTitleElement.textContent : 'Unknown';
        const artistName = artistNameElement ? artistNameElement.textContent : 'Unknown';
        const trackLink = trackLinkElement ? trackLinkElement.href : 'Unknown';
        const duration = durationElement ? durationElement.textContent.trim() : '05:31';
        const artworkUrl = artworkUrlElement ? artworkUrlElement.getAttribute('content') : '';
         const releaseDate = releaseDateElement ? releaseDateElement.getAttribute('datetime') : '';
        const tags = Array.from(tagElements).map(tag => tag.textContent.trim());

        console.log(artworkUrl)
        return { trackTitle, artistName, trackLink, duration, artworkUrl, releaseDate, tags};
    }

    // Отправка данных на сервер
    function sendTrackInfoToServer(trackInfo) {
        fetch('http://localhost:8080/api/submit', {
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
        sendTrackInfoToServer(trackInfo);
    }, 10000);

})();
```

## Layout

The layout functionality in this project allows users to create and manage customizable layouts. Here's an overview of how it works:

### Structure

- `UserLayout`: Represents a user's layout, including background, layout elements, creator, and editors.
- `LayoutElement`: Represents an individual element in the layout, with properties like position, type, and visibility.

### Permissions

- Creator: The user who created the layout. Has full edit permissions.
- Editors: Users who have been granted permission to edit the layout.
- Viewers: All other users who can view the layout but cannot edit it.

### Public vs Private Elements

- Each `LayoutElement` has a `Public` boolean field.
- If `Public` is true, the element is visible to all users.
- If `Public` is false, the element is only visible to the creator and editors.

### Key Operations

1. **GetUserLayout**: 
   - Retrieves a user's layout.
   - Filters out private elements if the requester doesn't have edit permissions.

2. **UpdateLayoutFull**: 
   - Updates the entire layout.
   - Requires edit permissions.
   - Logs the change.

3. **AddEditor**: 
   - Adds a new editor to the layout.
   - Only the creator can perform this action.
   - Logs the change.

4. **RemoveEditor**: 
   - Removes an editor from the layout.
   - Only the creator can perform this action.
   - Logs the change.

5. **IsEditor**: 
   - Checks if a user has edit permissions for a specific layout.

### Logging

All significant changes to layouts (updates, adding/removing editors) are logged using the `LogLayoutChange` method. This helps in tracking the history of changes made to layouts.

### Best Practices

- Always check permissions before allowing edits to a layout.
- Use the `Public` field to control the visibility of sensitive layout elements.
- Log all significant changes to maintain an audit trail.

For more detailed information on each operation, refer to the method documentation in the `layout` package.
