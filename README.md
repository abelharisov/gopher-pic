# gopher-pic

Приложение для конкурса https://t.me/omp_ru_gophercon2020

Приложение принимает на вход два изображения - картинку с человеком и картинку с  гофером и пытается их объеденить.

``` 
brew install opencv
brew install pkgconfig
go build cmd/tool/main.go

./main ./images/src.jpg ./images/gopher2.png ./images/out.png
``` 
