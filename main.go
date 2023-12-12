package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

type PosnMessage struct {
	Posn    string `json:"posn"`
	Message string `json:"message"`
}

type NilAway struct {
	Nilaway []PosnMessage `json:"nilaway"`
}

type finding struct {
	path       string
	message    string
	lineNumber int
	postition  int
	raw        PosnMessage
}

func main() {
	if data, err := os.ReadFile("nilaway.json"); err == nil {
		var report map[string]NilAway
		if err := json.Unmarshal(data, &report); err == nil {
			findings := map[string][]*finding{}
			if currentWorkDirectory, err := os.Getwd(); err != nil {
				panic(err)
			} else {
				if !strings.HasSuffix(currentWorkDirectory, "/") {
					currentWorkDirectory = currentWorkDirectory + "/"
				}
				for _, element := range report {
					for _, pm := range element.Nilaway {
						if f, err := findingFromPM(pm, currentWorkDirectory); err == nil {
							path := f.path
							if !strings.HasPrefix(path, ".go-cache/") {
								findings[path] = append(findings[path], f)
							}
						} else {
							panic(err)
						}
					}
				}
			}
			if file, err := os.Create("nilaway.checkstyle.xml"); err == nil {
				defer file.Close()
				file.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))
				file.Write([]byte(" <checkstyle version=\"5.0\">\n"))
				for path, pm := range findings {
					file.Write([]byte("  <file name=\""))
					file.Write([]byte(path))
					file.Write([]byte("\">\n"))
					for _, p := range pm {
						file.Write([]byte("   <error column=\""))
						file.Write([]byte(strconv.Itoa(p.postition)))
						file.Write([]byte("\" line=\""))
						file.Write([]byte(strconv.Itoa(p.lineNumber)))
						file.Write([]byte("\" message=\""))
						file.Write([]byte(clenMessage(p.message)))
						file.Write([]byte("\" severity=\"error\" source=\"nilaway\"/>\n"))
					}
					file.Write([]byte("  </file>\n"))

				}
				file.Write([]byte("</checkstyle>\n"))
				return
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}

}

func clenMessage(message string) string {
	m0 := strings.ReplaceAll(message, "\n", "&#xA;")
	m1 := strings.ReplaceAll(m0, "\r", "")
	m2 := strings.ReplaceAll(m1, "\u001B[0m", "")  //remove ansi color reset
	m3 := strings.ReplaceAll(m2, "\u001B[31m", "") //remove ansi color
	m4 := strings.ReplaceAll(m3, "\u001B[36m", "") //remove ansi color
	m5 := strings.ReplaceAll(m4, "\u001B[95m", "") //remove ansi color
	m6 := strings.ReplaceAll(m5, "\u003e", "")
	return strings.TrimPrefix(m6, "error: ")
}

func findingFromPM(pm PosnMessage, currentWorkDirectory string) (*finding, error) {
	message := pm.Message
	if path, line, pos, err := parsePosn(pm.Posn, currentWorkDirectory); err == nil {
		return &finding{message: message, path: path, lineNumber: line, postition: pos, raw: pm}, nil
	} else {
		return nil, err
	}
}

func parsePosn(posn string, currentWorkDirectory string) (string, int, int, error) {
	splited := strings.Split(posn, ":")
	path := splited[0]
	path = strings.TrimPrefix(path, currentWorkDirectory)
	if line, err := strconv.Atoi(splited[1]); err != nil {
		return path, -1, -1, err
	} else if column, err := strconv.Atoi(splited[2]); err != nil {
		return path, line, -1, err
	} else {
		return path, line, column, nil
	}
}
