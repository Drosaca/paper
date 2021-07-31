package qr

import (
	"fmt"
	rqrc "github.com/tuotoo/qrcode"
	wwqrc "github.com/yeqown/go-qrcode"
	"io/ioutil"
	"os"
	"paperify/stat"
	"path"
	"path/filepath"
	"strconv"
)

type Qr struct {
	ErrorCorrection uint // value from 0 to 3 (7%, 15%, 25%, 30%)
	InputPath       string
	OutputPath      string
	recipient       *os.File
	decodeStep     int
	chunks         map[int]string
}

func NewQr(inputPath string, outputPath string) Qr {
	return Qr{ErrorCorrection: 1, InputPath: inputPath, OutputPath: outputPath, recipient: nil, decodeStep: -1, chunks: map[int]string{}}
}

func (qr *Qr) generateQrCode(str string, filename string) error {
	fmt.Println("writing.. ", filename)
	config := wwqrc.Config{
		EcLevel: wwqrc.ErrorCorrectionLow,
		EncMode: wwqrc.EncModeByte,
	}
	qrc, err := wwqrc.NewWithConfig(str, &config)
	if err != nil {
		return err
	}
	err = qrc.Save(filename)
	return err
}

func (qr *Qr) CreateQr() error {
	f, err := os.Open(qr.InputPath)
	if err != nil {
		return err
	}
	buff := make([]byte, 2951) //b64 2213 default 2952
	part := 0
	for bytes, err := f.Read(buff); bytes != 0; {
		str := string(part) + string(buff[:bytes])
		fmt.Println(len(str))
		_, inputFileName := path.Split(qr.InputPath)
		err = qr.generateQrCode(str, path.Join(qr.OutputPath, inputFileName + fmt.Sprintf("_%d.png", part)))
		if err != nil {
			return err
		}
		for i := range buff {
			buff[i] = 0
		}
		bytes, err = f.Read(buff)
		if err != nil {
			return err
		}
		part += 1
	}
	return err
}

func (qr *Qr) ReadQr() error {
	var err error = nil
	if stat.IsDirectory(qr.OutputPath) {
		qr.recipient, err = os.Create(filepath.Join(qr.OutputPath, "output.raw"))
	} else {
		qr.recipient, err = os.Create(qr.OutputPath)
	}
	if stat.IsDirectory(qr.InputPath) {
		err = qr.readInDir()
	} else {
		err := qr.decodeQr(qr.InputPath)
		if err != nil {
			return err
		}
	}
	return err
}

func (qr *Qr) readInDir() error {
	err := filepath.Walk(qr.InputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		err = qr.decodeQr(path)
		if err != nil {
			return fmt.Errorf("fail to read Qr code in file %s  :%w", path, err)
		}
		return err
	})
	defer qr.recipient.Close()
	return err
}

func (qr *Qr) decodeQr(path string) error {
	fmt.Println("reading..", path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	qrCode, err := rqrc.Decode(file)
	if err != nil {
		return fmt.Errorf("Recognize failed: %v\n", err)
	}
	bytes := []byte(qrCode.Content)
	if int(bytes[0]) == qr.decodeStep + 1 {
		qr.decodeStep += 1
		_, err = qr.recipient.Write(bytes[1:])
	} else {
		err := qr.writeTmpChunk(bytes)
		if err != nil {
			return err
		}
	}
	err = qr.findStoredChunk()
	if err != nil {
		return err
	}
	return err
}

func (qr *Qr) writeTmpChunk(bytes []byte) error {
	fmt.Println("creating chunk")
	chunkPath := path.Join(os.TempDir(), strconv.Itoa(int(bytes[0])) + ".chunk")
	qr.chunks[int(bytes[0])] = chunkPath
	f, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	_, err = f.Write(bytes[1:])
	if err != nil {
		return err
	}
	return err
}

func (qr *Qr) findStoredChunk() error {
	for partNumber, fPath := range qr.chunks {
		if partNumber == qr.decodeStep + 1 {
			bytes, err := ioutil.ReadFile(fPath)
			if err != nil {
				return err
			}
			fmt.Println("chunk found : ",fPath, " merging..." )
			_, err = qr.recipient.Write(bytes[:])
			qr.decodeStep += 1
			err = qr.findStoredChunk()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

