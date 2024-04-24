#!/bin/bash

key=`/usr/bin/hexdump -vn8 -e'4/4 "%08X" 1 "\n"' /dev/urandom | sed 's/ //g'`
sed -i "s/const AESKEY = \"B5FFCEE73EF298A4\"/const AESKEY = \"$key\"/g" ./cipher.go
