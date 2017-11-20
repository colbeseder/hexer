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

type fileDetails struct {
    path string
    p_fileData  *[]byte
}

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

func saveFile(f fileDetails){
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

func saveFileAs(p_f *fileDetails){
    fn := getInput("File Name: ");
    (*p_f).path = fn;
    saveFile(*p_f);
}

func printFile(buf *[]byte , wait bool) {
    //fmt.Printf("file length: %d \n", len(*buf));
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

func truncate(fileBytes *[]byte){
    t := getInput("Truncate to length:  ");
    l, _ := strconv.ParseInt(t, 16, 64);
    newFileBytes := (*fileBytes)[:l];
    *fileBytes = newFileBytes;
}

func overwriteBytes(fileIdx int64, fileBytes *[]byte, newData []byte) error {
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

func hexStringToString(h string) string{
    reg, _ := regexp.Compile("[^a-fA-F0-9]");
    h = reg.ReplaceAllString(h, "");
    datAsBytes := []byte(h);
    buf := make([]byte, hex.DecodedLen(len(datAsBytes)));
    hex.Decode(buf, datAsBytes);
    return string(buf);
}

func replaceData(fileBytes *[]byte, isDataHex bool){
    t := getInput("Enter offset:  ");
    offset, e := strconv.ParseInt(t, 16, 64);
    if e != nil {
        panic(e);
    }
    dat := getInput("Enter data:  ");
    if (isDataHex){
        dat = hexStringToString(dat);
    }
    overwriteBytes(offset, fileBytes, []byte(dat));
}

func insertData(fileBytes *[]byte, isDataHex bool){
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
    overwriteBytes(offset + int64(len(datAsBytes)), fileBytes, toShift);
    overwriteBytes(offset, fileBytes, datAsBytes);
}

func appendData(fileBytes *[]byte, isDataHex bool){
    dat := getInput("Enter data:  ");
    if (isDataHex){
        dat = hexStringToString(dat);
    }
    overwriteBytes(int64(len(*fileBytes)), fileBytes, []byte(dat));
}

func newFilename(p_f *fileDetails){
    (*p_f).path = getInput("New Filename: ");
}

func runOption(selection string, p_f *fileDetails) error {
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
        replaceData(f.p_fileData, false);
        break;
    case "replace_hex" : fallthrough;
    case "rh" : fallthrough;
    case "edit_hex" : fallthrough;
    case "eh" :
        replaceData(f.p_fileData, true);
        break;
    case "print":  fallthrough;
    case "p":
        printFile(f.p_fileData, false);
        break;
    case "less":
        printFile(f.p_fileData, true);
    break;
    case "save": fallthrough;
    case "s":
        saveFile(f);
        break;
    case "save_as": fallthrough;
    case "sa":
        saveFileAs(p_f);
        break;
    case "truncate": fallthrough;
    case "trunc":
        truncate(f.p_fileData);
        break;
    case "append": fallthrough;
    case "a":
        appendData(f.p_fileData, false);
        break;
    case "append_hex": fallthrough;
    case "ah":
        appendData(f.p_fileData, true);
        break;
    case "insert": fallthrough;
    case "i":
        insertData(f.p_fileData, false);
        break;
    case "insert_hex": fallthrough;
    case "ih":
        insertData(f.p_fileData, true);
        break;
    case "quit": fallthrough;
    case "q":
        message("bye!");
        os.Exit(0);
        break;
    case "x":
        //Save and exit
        saveFile(f);
        message("Saved. bye!");
        os.Exit(0);
        break;
    default:
        return errors.New(selection + " is not a valid option");
    }
    return nil;
}

func getOption(f *fileDetails){
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

    fileBytes := make([]byte, 10);
    readFileToArray( path, &fileBytes );
    if (isInteractive && !isQuietStart) {
        printFile(&fileBytes, false);
    }
    f := fileDetails{path, &fileBytes} ;
    for {
        message("");
        getOption(&f);
    }
}
