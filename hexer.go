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


func message(str string){
    if isInteractive {
        fmt.Println(str);
    }
}

func log(s string){
    if isVerbose {
        fmt.Println(s);
    }
}

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

func (f doc) saveFile(){
    path := f.path;
    buf := f.p_fileData;
    if strings.TrimSpace(path) == "" {
        message("No filename given");
        return;
    }
    err := ioutil.WriteFile(path, *buf, 0644);
    if err != nil {
        message("Could not save");
        message(err.Error());
    }
}

func (f doc) saveFileAs(){
    fn := getInput("File Name: ");
    f.path = fn;
    f.saveFile();
}

func (f doc) printFile(wait bool) {
    buf := f.p_fileData;
    lineCount := int(math.Ceil( float64(len(*buf)) / float64(16)));
    for i := 0; i < lineCount ; i++  {
        endIdx := (i+1)*16;
        if endIdx > len(*buf) {
            endIdx = len(*buf);
        }
        l := (*buf)[i*16:endIdx];
        fmt.Println( formatLine(i, &l ));
        if wait {
            reader := bufio.NewReader(os.Stdin);
            text, _, _ := reader.ReadRune();
            if text == rune("q"[0]) {
                return;
            }
        }
    }
}

func (f doc) truncate(){
    fileBytes := f.p_fileData;
    t := getInput("Truncate to length:  ");
    l, _ := strconv.ParseInt(t, 16, 64);
    newFileBytes := (*fileBytes)[:l];
    *fileBytes = newFileBytes;
}

func (f doc) overwriteBytes(fileIdx int64, newData []byte) error {
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

func (f doc) replaceData(isDataHex bool){
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

func (f doc) deleteData(){
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

func (f doc) insertData(isDataHex bool){
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

func (f doc) appendData(isDataHex bool){
    fileBytes := f.p_fileData;
    dat := getInput("Enter data:  ");
    if (isDataHex){
        dat = hexStringToString(dat);
    }
    f.overwriteBytes(int64(len(*fileBytes)), []byte(dat));
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
    if (isInteractive){
        reader := bufio.NewReader(os.Stdin);
        fmt.Printf(msg);
        text, _ = reader.ReadString('\n');
        return trimTrailingLinebreaks(text);
    }
    if nextArg < len(os.Args) {
        text = os.Args[nextArg];
        log(">>> " + msg + " " + text);
    } else {
        // No more arguments
        text = "";
    }
    nextArg++ ;
    return text;
}

func hexStringToString(h string) string{
    reg, _ := regexp.Compile("[^a-fA-F0-9]");
    h = reg.ReplaceAllString(h, "");
    datAsBytes := []byte(h);
    buf := make([]byte, hex.DecodedLen(len(datAsBytes)));
    hex.Decode(buf, datAsBytes);
    return string(buf);
}

func runOption(selection string, p_f *doc) error {
    f := *p_f;
    switch selection {
    case "setVerbose":
        isVerbose = true;
        log("SET_VERBOSE");
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
    case "less":
        f.printFile(true);
    break;
    case "save": fallthrough;
    case "s":
        f.saveFile();
        break;
    case "save_as": fallthrough;
    case "sa":
        f.saveFileAs();
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
        message("bye!");
        os.Exit(0);
        break;
    case "x":
        //Save and exit
        f.saveFile();
        message("Saved. bye!");
        os.Exit(0);
        break;
    default:
        return errors.New(selection + " is not a valid option");
    }
    return nil;
}

func getOption(f *doc){
    op := getInput("Choose an option: ");
    if op == "" {
        op = "q";
    }
    e := runOption(op, f);
    if e != nil {
        fmt.Println(e);
    }
}

var nextArg int;
var isInteractive bool;
var isQuietStart bool;
var isVerbose bool = false;

func main() {
    path := os.Args[1];
    isInteractive = len(os.Args) < 3 || os.Args[2] != "-o";
    isQuietStart = len(os.Args) >= 3 && (os.Args[2] != "-q" || os.Args[2] != "-quiet");
    nextArg = 3;

    f := *(newDoc(path));
    
    if (isInteractive && !isQuietStart) {
        f.printFile(false);
    }
    for {
        message("");
        getOption(&f);
    }
}
