# Rexcel ![build](https://github.com/Lambels/rexcel/workflows/build/badge.svg)
Rexcel is a cli tool which identifies forumla cells which have a recursive formula. Cells which have a recursive formula arent allowed in excel.

## Installing
```
go install github.com/Lambels/rexcel@latest
```

## Using Rexcel
```
rexcel <fileName ...>
```
rexcel accepts an arbitrary amount of files passed to it.

running rexcel on the testdata:
```
rexcel ./testdata/ex1.xlsx ./testdata/ex2.xlsx ./testdata/ex3.xlsx ./testdata/ex4.xlsx
./testdata/ex1.xlsx
A6
B7
B6
A7
C1
./testdata/ex2.xlsx
F5
F9
F6
F7
F8
./testdata/ex3.xlsx
A1
B1
./testdata/ex4.xlsx
B1
A2
A1
```
the output represents the recursive cells for each excel file.