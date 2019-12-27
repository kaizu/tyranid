package main

import (
    "os"
    "bytes"
    "io"
    "fmt"
    "strconv"
    "strings"
    "net/http"

    "encoding/json"
    "github.com/go-yaml/yaml"

    "github.com/ktnyt/gods"
    "github.com/ktnyt/gt1"

    "./csv"
)

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

func parseOperon(row []string) (*Operon, error) {
    var operon Operon
    operon.Name = row[0]
    left, err := strconv.Atoi(row[1])
    if err != nil {
        return nil, err
    }
    operon.Left = left
    right, err := strconv.Atoi(row[1])
    if err != nil {
        return nil, err
    }
    operon.Right = right
    if row[3] == "forward" || row[3] == "reverse" {
        operon.Strand = (row[3] == "forward")
    } else {
        return nil, fmt.Errorf("Invbalid DNA strand was given [%s].", row[3])
    }
    n, err := strconv.Atoi(row[4])
    if err != nil {
        return nil, err
    }
    operon.NumOfGenes = n
    operon.GeneNames = strings.Split(row[5], ",")
    if n != len(operon.GeneNames) {
        return nil, fmt.Errorf("The number of gene names doesn't match [%d != %d]", n, len(operon.GeneNames))
    }
    operon.Evidence = row[6]
    operon.Confidence = row[7]
    return &operon, nil
}

type Scanner struct {
    scanner *csv.Scanner
    record *Operon
    err error
    continueOnError bool
}

func NewScanner(reader io.Reader) *Scanner {
    return new(Scanner).initialize(reader)
}

func (this *Scanner) initialize(reader io.Reader) *Scanner {
    this.scanner = csv.NewScanner(reader, csv.Comma('\t'), csv.Comment('#'), csv.FieldsPerRecord(8))
    return this
}

func (this *Scanner) Scan() bool {
    var suc bool = this.scanner.Scan()
    this.err = this.scanner.Error()
    if !suc {
        return false
    }

    if this.err != nil {
        return true
    }

    this.record, this.err = parseOperon(this.scanner.Record())
    return true
}

func (this *Scanner) Record() *Operon {
	return this.record
}

func (this *Scanner) Error() error {
	if this.err == io.EOF {
		return nil
	}
	return this.err
}

func parseFeature(operon *Operon) *gt1.Feature {
    var location gt1.Location = gt1.NewRangeLocation(operon.Left, operon.Right)
    if !operon.Strand {
        location = gt1.NewComplementLocation(location)
    }
    qfs := gods.NewOrdered()
    qfs.Add("operon", operon.Name)
    qfs.Add("db_xref", fmt.Sprintf("REGULONDB:%s", operon.Name))
    return gt1.NewFeature("operon", location, qfs)
}

// Almost same with gt1.featureIO
type FeatureIO struct {
    Key string `yaml:"key" json:"key"`
    Location string `yaml:"location" json:"location"`
    Qualifiers [][2]string `yaml:"qualifiers" json:"qualifiers"`
}

func NewFeatureIO(feature *gt1.Feature) *FeatureIO {
    qfs := make([][2]string, feature.Qualifiers().Len())
    for i, item := range feature.Qualifiers().Iter() {
        qfs[i][0] = item.Key
        qfs[i][1] = item.Value
    }
    return &FeatureIO{feature.Key(), feature.Location().Format(), qfs}
}

func formatFeatureJson(feature *gt1.Feature) (string, error) {
    fmt := NewFeatureIO(feature)

    buf, err := json.Marshal([]*FeatureIO{fmt})
    if err != nil {
        return "", err
    }

    var newbuf bytes.Buffer
    err = json.Indent(&newbuf, buf, "", "  ")
    if err != nil {
        return "", err
    }
    return newbuf.String(), nil
}

func formatFeatureYaml(feature *gt1.Feature) (string, error) {
    fmt := NewFeatureIO(feature)

    buf, err := yaml.Marshal([]*FeatureIO{fmt})
    if err != nil {
        return "", err
    }
    return string(buf), nil
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

    scanner := NewScanner(file)
    for {
        if suc := scanner.Scan(); !suc {
            break
        }
        if err = scanner.Error(); err != nil {
            continue
        }
        // res, err := formatFeatureYaml(parseFeature(scanner.Record()))
        res, err := formatFeatureJson(parseFeature(scanner.Record()))
        if err != nil {
            panic(err)
        }
        fmt.Println(res)
    }
}
