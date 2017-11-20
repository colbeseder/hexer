# hexer

## A cmd hex editor

### Interactive mode

#### Start editing:
    $ ./hexer file.txt

#### Start editing

    00000000   54 68 69 73 20 69 73 20 6d 79 20 70 72 61 63 74    This.is.my.pract
    00000010   69 63 65 20 66 69 6c 65                            ice.file
    
    Choose an option: insert
    Enter offset:  0b
    Enter data:  great 
    
    Choose an option: print
    00000000   54 68 69 73 20 69 73 20 6d 79 20 67 72 65 61 74    This.is.my.great
    00000010   20 70 72 61 63 74 69 63 65 20 66 69 6c 65          .practice.file

    Choose an option: save_as
    File Name: new_file.txt

    Choose an option: edit_hex
    Enter offset:  0b
    Enter data:  73 75 70 65 72

    Choose an option: print
    00000000   54 68 69 73 20 69 73 20 6d 79 20 73 75 70 65 72    This.is.my.super
    00000010   20 70 72 61 63 74 69 63 65 20 66 69 6c 65          .practice.file

    Choose an option: save

    Choose an option: quit


### Automated mode (_-o_)

Pass in all commands at start

    $ ./hexer file.txt -o insert 0b "great " save_as new_file.txt

Full command list:

####edit
(aka **e**, **replace**, **r**)

Write text over existing bite, starting from provided offset.

####edit_hex
(aka **eh**, **replace_hex**, **rh**)

Write bytes over existing bytes starting from provided offset. Bytes should be provided in hex format. Non-hex characters (eg. spaces) will be ignored.

####print
(aka **p**)

Send hex-formatted file to stdout

####less
Send hex-formatted file to stdout one line at a time and wait for return to be pressed (even in automated mode)

####save
(aka **s**)

Save edited file to disk

####save_as
(aka **sa**)

Save file to disk with new name

####truncate
(aka **trunc**)

Delete all bytes from provided offset

####append
(aka **a**)
Add text to end of file

####append_hex
(aka **ah**)

Add bytes to end of file. Bytes should be provided in hex format

####insert
(aka **i**)

Add text to file at provided offset

####insert_hex
(aka **ih**)

Add bytes to file at provided offset. Bytes should be provided in hex format

####delete
(aka **d**)

Delete specified number of bytes, starting at the offset

####quit
(aka **q**)

Exit _hexer_

####x
Save and exit