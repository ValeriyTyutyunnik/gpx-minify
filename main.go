package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
)

// Описываем структуру GPX, включая только те поля, которые нужно оставить
type GPX struct {
	XMLName        xml.Name `xml:"http://www.topografix.com/GPX/1/1 gpx"`
	Creator        string   `xml:"creator,attr,omitempty"`
	Version        string   `xml:"version,attr,omitempty"`
	XSI            string   `xml:"xmlns:xsi,attr,omitempty"`
	SchemaLocation string   `xml:"xsi:schemaLocation,attr,omitempty"`
	GPXTPX         string   `xml:"xmlns:gpxtpx,attr,omitempty"`
	GPXX           string   `xml:"xmlns:gpxx,attr,omitempty"`
	Metadata       Metadata `xml:"metadata"`
	Trk            Track    `xml:"trk"`
}

type Metadata struct {
	Name string `xml:"name,omitempty"`
	Desc string `xml:"desc,omitempty"`
	Time string `xml:"time,omitempty"`
}

type Track struct {
	Name   string   `xml:"name,omitempty"`
	Type   string   `xml:"type,omitempty"`
	TrkSeg TrackSeg `xml:"trkseg"`
}

type TrackSeg struct {
	TrkPt []TrackPt `xml:"trkpt"`
}

type TrackPt struct {
	Lat  string `xml:"lat,attr"`
	Lon  string `xml:"lon,attr"`
	Ele  string `xml:"ele,omitempty"`
	Time string `xml:"time,omitempty"`
	// Другие поля умышленно не добавляем в структуру, чтобы просто проигнорировать и не сохранять их
}

func formatBytes(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMG"[exp])
}

func main() {
	inputPtr := flag.String("in", "", "Путь к исходному GPX файлу (можно полный путь)")
	outputPtr := flag.String("out", "output.gpx", "Путь для сохранения очищенного GPX файла")
	flag.Parse()

	inputPath := *inputPtr
	outputPath := *outputPtr

	if inputPath == "" {
		fmt.Println("GPX файл не указан")
	}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Printf("Ошибка открытия файла: %v\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	// Читаем данные из файла целиком
	byteValue, err := io.ReadAll(inputFile)
	if err != nil {
		fmt.Printf("Ошибка чтения файла: %v\n", err)
		os.Exit(1)
	}

	// Десериализуем XML в структуры
	var gpx GPX
	if err := xml.Unmarshal(byteValue, &gpx); err != nil {
		fmt.Printf("Файл должен содержать GPX трек. Ошибка парсинга XML: %v\n", err)
		os.Exit(1)
	}

	// Сериализуем структуру обратно в XML с отступами
	outputBytes, err := xml.MarshalIndent(gpx, "", " ")
	if err != nil {
		fmt.Printf("Ошибка создания XML: %v\n", err)
		os.Exit(1)
	}

	// Добавляем стандартный XML-заголовок
	xmlHeader := []byte(xml.Header)
	finalOutput := append(xmlHeader, outputBytes...)

	fmt.Println(formatBytes(len(byteValue)), formatBytes(len(finalOutput)))

	// Ничего не сделали
	if len(byteValue) == len(finalOutput) {
		fmt.Println("Нечего сжимать")
		os.Exit(0)
	}

	err = os.WriteFile(outputPath, finalOutput, 0644)
	if err != nil {
		fmt.Printf("Ошибка записи: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Трек успешно очищен в %s. Первоначальный размер=%s, Новый размер=%s", outputPath, formatBytes(len(byteValue)), formatBytes(len(finalOutput)))
	os.Exit(0)
}
