package main

import (
    "os"
    "fmt"
    "errors"
    "math"
    "io/ioutil"
    "encoding/hex"
    "regexp"
    "strings"
    "bufio"
    "strconv"
)

var v viewer;

func leftPad(s string, padStr string, pLen int) string {
    return strings.Repeat(padStr, pLen - len(s)) + s;
}

func readFileToArray(path string, buf *[]byte ) {
    *buf = make([]byte, 0);
    if strings.TrimSpace(path) == "" {
        return;
    }
    dat, err := ioutil.ReadFile(path);
    if os.IsNotExist(err){
        return;
    } else if err != nil {
        panic(err);
    }
    *buf = dat;
}

type doc struct {
    path string
    p_fileData  *[]byte
}

func (f *doc) saveFile(p_v *viewer){
    v := *p_v;
    path := f.path;
    buf := f.p_fileData;
    if strings.TrimSpace(path) == "" {
        v.message("No filename given");
        return;
    }
    err := ioutil.WriteFile(path, *buf, 0644);
    if err != nil {
        v.message("Could not save");
        v.message(err.Error());
    }
}

func (f *doc) saveFileAs(p_v *viewer){
    fn := getInput("File Name: ");
    f.path = fn;
    f.saveFile(p_v);
}

func (f *doc) printFile(wait bool) {
    f.printFileSection(wait, 0, 0);
}

func (f *doc) printFileSection(wait bool, start int, stop int) {
    buf := f.p_fileData;
    lineCount := int(math.Ceil( float64(len(*buf)) / float64(16)));
    if stop == 0 || stop > lineCount{
        stop = lineCount;
    }

    startLine := start / 16;
    
    for line := startLine; line < stop ; line++  {
        endIdx := (line+1)*16;
        if endIdx > len(*buf) {
            endIdx = len(*buf);
        }
        l := (*buf)[line*16:endIdx];
        fmt.Println( formatLine(line, &l ));
        if wait {
            reader := bufio.NewReader(os.Stdin);
            text, _, _ := reader.ReadRune();
            if text == rune("q"[0]) {
                return;
            }
        }
    }
}

func (f *doc) truncate(){
    fileBytes := f.p_fileData;
    t := getInput("Truncate to length:  ");
    l, _ := strconv.ParseInt(t, 16, 64);
    newFileBytes := (*fileBytes)[:l];
    *fileBytes = newFileBytes;
}

func (f *doc) overwriteBytes(fileIdx int64, newData []byte) error {
    fileBytes := f.p_fileData;
    dataIdx := 0;
    dataLen := len(newData);
    fileLen := int64(len(*fileBytes));

    if fileIdx + int64(dataLen) > fileLen {
        //Make file byte array long enough to accept the data
        bytesToAdd := fileIdx + int64(dataLen) - fileLen;
        newFileBytes := append(*fileBytes, make([]byte, bytesToAdd ) ...);
        *fileBytes = newFileBytes;
        fileLen = int64(len(*fileBytes));
    }

    for dataIdx < dataLen && fileIdx < fileLen {
        (*fileBytes)[fileIdx] = newData[dataIdx];
        dataIdx++;
        fileIdx++;
    }

    return nil;
}

func (f *doc) replaceData(isDataHex bool){
    t := getInput("Enter offset:  ");
    offset, e := strconv.ParseInt(t, 16, 64);
    if e != nil {
        panic(e);
    }
    dat := getInput("Enter data:  ");
    if (isDataHex){
        dat = hexStringToString(dat);
    }
    f.overwriteBytes(offset, []byte(dat));
}

func (f *doc) deleteData(){
    fileBytes := f.p_fileData;
    t := getInput("Enter offset:  ");
    offset, e := strconv.ParseInt(t, 16, 64);
    l, e := strconv.ParseInt(getInput("Enter number of bytes:  "), 16, 64);

    if e != nil {
        panic(e);
    }

    toShift := make([]byte, int64(len(*fileBytes)) - (offset + l));
    copy(toShift, (*fileBytes)[offset+l:]);
    f.overwriteBytes(offset, toShift);

    newFileLength := int64(len(*fileBytes)) - l;

    newFileBytes := (*fileBytes)[:newFileLength];
    *fileBytes = newFileBytes;
}

func (f *doc) insertData(isDataHex bool){
    fileBytes := f.p_fileData;
    t := getInput("Enter offset:  ");
    offset, e := strconv.ParseInt(t, 16, 64);
    if e != nil {
        panic(e);
    }
    dat := getInput("Enter data:  ");
    if (isDataHex){
        dat = hexStringToString(dat);
    }
    datAsBytes := []byte(dat);
    toShift := make([]byte, int64(len(*fileBytes)) - offset);
    copy(toShift, (*fileBytes)[offset:]);
    f.overwriteBytes(offset + int64(len(datAsBytes)), toShift);
    f.overwriteBytes(offset, datAsBytes);
}

func (f *doc) appendData(isDataHex bool){
    fileBytes := f.p_fileData;
    dat := getInput("Enter data:  ");
    if (isDataHex){
        dat = hexStringToString(dat);
    }
    f.overwriteBytes(int64(len(*fileBytes)), []byte(dat));
}

type viewer struct {
    args []string
    isInteractive bool
    isQuietStart bool
    isVerbose bool
    nextArg int
}

func (v *viewer) getNextArg(msg string) string {
    r := ""
    if v.nextArg < len(v.args) {
        r = v.args[v.nextArg];
    }
    v.log(">>> " + msg + " " + r);
    v.nextArg++ ;
    return r;
}

func (v *viewer) message(str string){
    if v.isInteractive {
        fmt.Println(str);
    }
}

func (v viewer) log(str string){
    if v.isVerbose {
        fmt.Println(str);
    }
}


func newDoc(path string) *doc {
    fileBytes := make([]byte, 10);
    readFileToArray( path, &fileBytes );
    f := doc{path, &fileBytes} ;
    return &f;
}

func formatLine(lineNum int, line *[]byte) string{
    l := len(*line);
    reg, _ := regexp.Compile("[^!-~]");
    offset := lineNum * 16;
    buf := make([]byte, hex.EncodedLen(l));
    hex.Encode(buf, *line);
    r := make([]string, 16);
    for i := 0 ; i < 16 ; i++ {
        if i < l {
            r[i] = string(buf[i*2:(i*2)+2]);
        } else {
            r[i] = "  ";
        }
    }

    hexChars := strings.Join(r, " ");

    offsetString := leftPad(fmt.Sprintf("%x", offset), "0", 8);
    return offsetString + "   " + hexChars + "    " + reg.ReplaceAllString(string(*line), ".");
}

func trimTrailingLinebreaks(txt string) string {
    reg, _ := regexp.Compile("[\\n\\r]+$");
    return reg.ReplaceAllString(txt, "");
}

func getInput(msg string) string {
    var text string;
    if (v.isInteractive){
        reader := bufio.NewReader(os.Stdin);
        fmt.Printf(msg);
        text, _ = reader.ReadString('\n');
        return trimTrailingLinebreaks(text);
    }
    return v.getNextArg(msg);
}

func hexStringToString(h string) string{
    reg, _ := regexp.Compile("[^a-fA-F0-9]");
    h = reg.ReplaceAllString(h, "");
    datAsBytes := []byte(h);
    buf := make([]byte, hex.DecodedLen(len(datAsBytes)));
    hex.Decode(buf, datAsBytes);
    return string(buf);
}

func runOption(selection string, p_f *doc, p_v *viewer) error {
    f := *p_f;
    v := *p_v;
    switch selection {
    case "setVerbose":
        v.isVerbose = true;
        v.log("SET_VERBOSE");
        break;
    case "replace" : fallthrough;
    case "r" : fallthrough;
    case "edit" : fallthrough;
    case "e" :
        f.replaceData(false);
        break;
    case "delete" : fallthrough;
    case "d" :
        f.deleteData();
        break;
    case "replace_hex" : fallthrough;
    case "rh" : fallthrough;
    case "edit_hex" : fallthrough;
    case "eh" :
        f.replaceData(true);
        break;
    case "print":  fallthrough;
    case "p":
        f.printFile(false);
        break;
    case "print_segment":  fallthrough;
    case "ps":
        startOffset, _ := strconv.ParseInt(getInput("Start Offset: "), 16, 64);
        f.printFileSection(true, int(startOffset) , 0);
        break;
    case "less":
        f.printFile(true);
    break;
    case "save": fallthrough;
    case "s":
        f.saveFile(p_v);
        break;
    case "save_as": fallthrough;
    case "sa":
        f.saveFileAs(p_v);
        break;
    case "truncate": fallthrough;
    case "trunc":
        f.truncate();
        break;
    case "append": fallthrough;
    case "a":
        f.appendData(false);
        break;
    case "append_hex": fallthrough;
    case "ah":
        f.appendData(true);
        break;
    case "insert": fallthrough;
    case "i":
        f.insertData(false);
        break;
    case "insert_hex": fallthrough;
    case "ih":
        f.insertData(true);
        break;
    case "quit": fallthrough;
    case "q":
        v.message("bye!");
        os.Exit(0);
        break;
    case "x":
        //Save and exit
        f.saveFile(p_v);
        v.message("Saved. bye!");
        os.Exit(0);
        break;
    default:
        return errors.New(selection + " is not a valid option");
    }
    return nil;
}

func getOption(f *doc, p_v *viewer){
    op := getInput("Choose an option: ");
    if op == "" {
        op = "q";
    }
    e := runOption(op, f, p_v);
    if e != nil {
        fmt.Println(e);
    }
}

const IS_VERBOSE = false;

func main() {
    path := os.Args[1];
    isInteractive := len(os.Args) < 3 || os.Args[2] != "-o";
    isQuietStart := len(os.Args) >= 3 && (os.Args[2] != "-q" || os.Args[2] != "-quiet");
    v = viewer{ os.Args, isInteractive, isQuietStart, IS_VERBOSE, 3 } ;

    f := *(newDoc(path));
    
    if (v.isInteractive && !v.isQuietStart) {
        f.printFile(false);
    }
    for {
        v.message("");
        getOption(&f, &v);
    }
}
