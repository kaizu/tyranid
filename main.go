package main

import (
    "encoding/csv"
    "os"
    "io"
    "fmt"
    "strconv"
    "strings"
    "net/http"
)

// (1) Operon name
// (2) First gene-position left
// (3) Last gene-position right
// (4) DNA strand where the operon is coded
// (5) Number of genes contained in the operon
// (6) Name or Blattner number of the gene(s) contained in the operon
// (7) Evidence that support the existence of the operon's TUs
// (8) Evidence confidence level (Confirmed, Strong, Weak)

type Operon struct {
    Name string
    Left int
    Right int
    Strand bool
    NumOfGenes int
    GeneNames []string
    Evidence string
    Confidence string
}

func parseOperon(row []string) (Operon, error) {
    var operon Operon
    operon.Name = row[0]
    left, err := strconv.Atoi(row[1])
    if err != nil {
        return Operon{}, err
    }
    operon.Left = left
    right, err := strconv.Atoi(row[1])
    if err != nil {
        return Operon{}, err
    }
    operon.Right = right
    if row[3] == "forward" || row[3] == "reverse" {
        operon.Strand = (row[3] == "forward")
    } else {
        return Operon{}, fmt.Errorf("Invbalid DNA strand was given [%s].", row[3])
    }
    n, err := strconv.Atoi(row[4])
    if err != nil {
        return Operon{}, err
    }
    operon.NumOfGenes = n
    operon.GeneNames = strings.Split(row[5], ",")
    if n != len(operon.GeneNames) {
        return Operon{}, fmt.Errorf("The number of gene names doesn't match [%d != %d]", n, len(operon.GeneNames))
    }
    operon.Evidence = row[6]
    operon.Confidence = row[7]
    return operon, nil
}

func generateFeatureYaml(operon *Operon) string {
    location := fmt.Sprintf("%d..%d", operon.Left, operon.Right)
    if !operon.Strand {
        location = fmt.Sprintf("complement(%s)", location)
    }
    return fmt.Sprintf("- key: operon\n  location: %s\n  operon: %s\n  qualifiers:\n  - - db_xref\n    - REGULONDB:%s\n", location, operon.Name, operon.Name)
}

func isExist(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}

func fetchURL(filename string, url string) error {
    if isExist(filename) {
        return nil
    }

    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}

func main() {
    err := fetchURL("OperonSet.txt", "http://regulondb.ccg.unam.mx/menu/download/datasets/files/OperonSet.txt")
    if err != nil {
        panic(err)
    }

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

    for {
        line, err = reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        if operon, err := parseOperon(line); err == nil {
            fmt.Printf(generateFeatureYaml(&operon))
        }
    }
}
