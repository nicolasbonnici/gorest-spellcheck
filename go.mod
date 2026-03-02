module github.com/nicolasbonnici/gorest-spellcheck

go 1.26

toolchain go1.26.0

require (
	github.com/gofiber/fiber/v2 v2.52.12
	github.com/google/uuid v1.6.0
	github.com/nicolasbonnici/gorest v0.4.12
)

// Use local gorest for development
replace github.com/nicolasbonnici/gorest => ../gorest

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.8.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.4 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.69.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
