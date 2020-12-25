#!/bin/bash
echo it is sv.sh: that is setting version of makcommon library

version="+++da21c61--*main--201225_0801+++"
fix=+++
version=$(echo "$version" | sed -e "s/^$fix//" -e "s/$fix$//")
echo version=$version
#It is assuming that this variable is set into version of a host project and have format ---da21c61--*main--201225_0612---
#where
#A host project is that project which used the library
#PN is the host project name, e.g. 201216_rels
#VER is the host project version, e.g. da21c61--*main--201224_1311

#Also it is assuming that this script will be run 
#exit






{
sed -i "s/---.*---/---$version---/" ksess/rules.txt
sed -i "s/---.*---/---$version---/" ksess/api.txt
sed -i "s/---.*---/---$version---/" ksess/version.go
}


echo $1
#exit
#201223 05:15 does not roll
#exit

git add .
echo passed git add . 

git commit -m "$1"
echo git commit -m "$1"


git push

echo sv.sh: successfully ended ------------------------------------------------------








