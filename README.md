# patcher

#### description

This is a desktop program based on Go and Fyne.

step 1.
the module `imetin2/patcher/admin` is used to generate the MD5 configuration file.
Then you can upload the client files along with the configuration file to the web server.

step 2.
the module `imetin2/patcher` is the pather program.
You just need to configure the constant in main.go.

#### build

**go build without cli window**

`go build  -ldflags -H=windowsgui main.go`

**fyne build（windows）**

`fyne package -os windows -icon metin2.ico`


