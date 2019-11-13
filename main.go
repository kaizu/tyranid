package main

import (
    "encoding/csv"
    "os"
    "io"
    "fmt"
)

func GenerateFeatureYaml(row []string) (string, error) {
    location := fmt.Sprintf("%s..%s", row[1], row[2])
    if row[3] == "reverse" {
        location = fmt.Sprintf("complement(%s)", location)
    } else if row[3] != "forward" {
        return "", fmt.Errorf("Invbalid DNA strand was given [%s].", row[3])
    }
    yaml := fmt.Sprintf("- key: operon\n  location: %s\n  operon: %s\n  qualifiers:\n  - - db_xref\n    - REGULONDB:%s\n", location, row[0], row[0])
    return yaml, nil
}

func main() {
    file, err := os.Open("OperonSet.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()


    reader := csv.NewReader(file)
    reader.Comma = '\t'
    reader.Comment = '#'
    reader.FieldsPerRecord = 8

    var line []string
    var yml string

    for {
        line, err = reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        yml, err = GenerateFeatureYaml(line)
        if err == nil {
            fmt.Printf(yml)
        }
    }
}
