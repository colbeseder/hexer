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
    dat, err := ioutil.ReadFile(path);
    if os.IsNotExist(err){
        *buf = make([]byte, 0);
        return;
    } else if err != nil {
        panic(err);
    }
    *buf = dat;
}

func saveFile(path string, buf *[]byte){
    err := ioutil.WriteFile(path, *buf, 0644);
    if err != nil {
        message("Could not save");
        message(err.Error());
    }
}

func printFile(buf *[]byte ) {
    //fmt.Printf("file length: %d \n", len(*buf));
    lineCount := int(math.Ceil( float64(len(*buf)) / float64(16)));
    for i := 0; i < lineCount ; i++  {
        endIdx := (i+1)*16;
        if endIdx > len(*buf) {
            endIdx = len(*buf);
        }
        l := (*buf)[i*16:endIdx];
        fmt.Printf( formatLine(i, &l ));
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
    return offsetString + "   " + hexChars + "    " + reg.ReplaceAllString(string(*line), ".") + "\n";
}

func getInput(msg string) string {
    var text string;
    if (isInteractive){
        reader := bufio.NewReader(os.Stdin);
        fmt.Printf(msg);    
        text, _ = reader.ReadString('\n');
        return strings.TrimSpace(text);        
    }
    if nextArg < len(os.Args) {
        text = os.Args[nextArg];
        log(">>> " + msg + " " + text);
    } else {
        // No more arguments
        text =  "";
    }
    nextArg++ ;
    return strings.TrimSpace(text);
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
    datAsBytes :=  []byte(h);
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

func appendData(fileBytes *[]byte, isDataHex bool){
    dat := getInput("Enter data:  ");
    if (isDataHex){
        dat = hexStringToString(dat);
    }
    overwriteBytes(int64(len(*fileBytes)), fileBytes, []byte(dat));
}

func runOption(selection string, f fileDetails) error {
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
        printFile(f.p_fileData);
        break;
    case "save": fallthrough;
    case "s":
        saveFile(f.path, f.p_fileData);
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
        log("WARNING: 'insert' is not implemented");
        break;
    case "quit": fallthrough;
    case "q":
        message("bye!");
        os.Exit(0);
        break;
    case "x":
        //Save and exit
        saveFile(f.path, f.p_fileData);
        message("Saved. bye!");
        os.Exit(0);
        break;
    default:
        return errors.New(selection + " is not a valid option");
    }
    return nil;
}

func getOption(f fileDetails){
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
var isVerbose bool = false;

func main() {
    path := os.Args[1];
    isInteractive = len(os.Args) < 3 || os.Args[2] != "-o";
    nextArg = 3;

    fileBytes := make([]byte, 10);
    readFileToArray( path, &fileBytes );
    if (isInteractive) {
        printFile(&fileBytes);
    }
    f := fileDetails{path, &fileBytes} ;
    for {
        message("");
        getOption(f);
    }
}