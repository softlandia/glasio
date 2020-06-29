# golang las library #

(c) softlandia@gmail.com

>download: go get -u github.com/softlandia/glasio  
>install: go install

The library makes it easy to read or write data in LAS format.
The main reason for the development was the need to read and bring in a uniform form a large number of LAS files obtained from various sources

Features:

1. The encoding is determined automatically
2. On reading performed validation of the key parameters and is integrity of the structure LAS file
3. Messages are generated for all inconsistencies:
    - zero value of important parameters
    - depth step change in data section
    - lack of curves section
    - conversion errors to a numerical value
    - mismatch of the step parameter declared in the header to the actual
    - duplication of curve names
4. Excluding critical errors, the library allows you to read the file and get data
5. Saving a file ensures the integrity of the structure and its readability for most other programs
6. It is possible to specify a dictionary of standard mnemonics; when reading a file, messages about curves that do not match the specified ones will be generated
7. It is possible to specify a dictionary of automatic substitution of mnemonics, respectively, curves with the given names will be renamed

__WRAP__ las file not support

## dependences ##

- github.com/softlandia/cpd
- github.com/softlandia/xlib

## examples ##

simple

- make empty LAS file
- reads sample file "expand_points_01.las", write md file with messages
- saves the recovered LAS file "expand_points_01+.las"

repaire

- reads all LAS files in current folder
- saves the recovered files to the same folder

lasin

- reads LAS file
- print warning

## tests ##

coverage 91%  
folder "data" contain files for testing, no remove/change/add

## technical info ##

### how type Las store data ###

access to main parameters:  
las.VERS()  
las.WRAP()  
las.STEP()  
las.STRT()  
las.STOP()  
las.NULL()  
las.WELL()

number of points and curves:  
las.NumPoints() - number of points
len(las.Logs) - number of curves

access to curves and data:  
las.Logs[0].D[0] - first depth  
las.Logs[1].V[100] - value of first curve on 101 depth step  
las.Logs[2].Name - name of second curve  
las.Logs[2].Unit - unit of second curve  
las.Logs[2].Mnemonic - mnemonic of second curve, the value is determined if the dictionary was applied  

if las file contane duplicated of any parameter, then used first
on curve section used all curves name, duplicated renamed

## warnings generated when reading a LAS file ##

### warning format ###

extended:  
> x, y, "message text"  
> x - section number  
> y - line number of input file  

short:  
