# Paper
Go implementation of [paperify](https://github.com/alisinabh/paperify)


## Usage
```
Usage of paper:
-digital
aka digitalify -> reads the Qr code and writes THE FILE to output path
-i string
input file path
-o string
output path (default ".")
-paper
aka paperify -> creates the Qr code from ONE FILE
-v    verbose
```

```
Example:

$> paperify -i /path/to/file -o /path/to/output_folder -paper 

$> paperify -i /path/to/qrcodes/folder -o ./filename.file -digital
```
