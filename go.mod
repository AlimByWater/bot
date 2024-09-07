module arimadj-helper

go 1.21.4

require (
	github.com/PuerkitoBio/goquery v1.9.2
	github.com/antchfx/htmlquery v1.3.2
	github.com/bogem/id3v2 v0.0.0-00010101000000-000000000000
	github.com/essentialkaos/go-icecast v2.0.5+incompatible
	github.com/fatih/color v1.17.0
	github.com/gin-gonic/gin v1.10.0
	github.com/go-co-op/gocron v1.37.0
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.2-0.20221020003552-4126fa611266
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/websocket v1.5.3
	github.com/grafov/m3u8 v0.12.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/mailru/easyjson v0.7.7
	github.com/redis/go-redis/v9 v9.6.0
	github.com/schollz/progressbar/v3 v3.14.4
	github.com/stretchr/testify v1.9.0
	github.com/telegram-mini-apps/init-data-golang v1.1.5
	github.com/torden/go-strutil v0.1.7
	golang.org/x/net v0.26.0
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/antchfx/xpath v1.3.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.55.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/term v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	pkg.re/essentialkaos/check.v1 v1.2.0 // indirect
)

replace github.com/go-telegram-bot-api/telegram-bot-api/v5 => ./pkg/telegram-bot-api/

replace github.com/bogem/id3v2 => ./pkg/id3v2/
