# hexer

##Usage:

###Interactive mode

    ./hexer file.txt 

###CMD mode

Write over file starting from offset zero. Then save.

    ./hexer file.txt -o replace 0 "hello" save

Add hex bytes to end of file and then write hex viewer output to stdout (don't save)

    ./hexer file.txt -o append_hex "4c6f72656d" print

