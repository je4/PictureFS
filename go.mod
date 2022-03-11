module github.com/je4/PictureFS/v2

replace (
	github.com/je4/PictureFS/v2 => ./
)

go 1.17

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/Rayleigh865/gopack v0.0.0-20200510061658-cebe9d11a05e
	github.com/disintegration/imaging v1.6.2
	github.com/pkg/errors v0.9.1
)

require golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8 // indirect
