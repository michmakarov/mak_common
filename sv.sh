#!/bin/bash

#201228 12:37
#The idea: 
#A version of a library is the version of a project that modified the library last time.
#For achieving that this script must be invoked by a script that builds the project - let's it is named project script.
#The project script must pass the project version to here as first parameter.

#The project script must take care to change the working directory before invoking this script

#201230 08:12 In pursuing the xhrboo second parameter was allowed - toGit
# if it passed changes will be writen to the git

#+++++++++++++++++++++++++++++++++++++++++
#210423 05:41
#It is for saving to the git (with pushing to github) the library mak_kommon
#Also it is tagging with version the version.go files of the libraty components and non golang files as .html, .js and so on
#It assumes that it is secondary script, that should be invoked by some main script building a golang project.
#For details see comments into mak_common/mversion/ecxec.go
#From the main script this expects the version as first parameter and the demand to git the library as second parameter

goOutOnError(){
	local lastRetCode=$?
	local operName="$1"
	if [ -z "$operName" ]; then {
		operName="unknown operation"
	}
	fi

	if [ $lastRetCode != 0 ]; then {
		echo "Error of executing $operName(retCode:$lastRetCode)";
		exit
	} else { echo "--- $operName perfomed"; }
	fi
}

echo it is sv.sh: that is git committing and pushing mak_common library


areChanges=$(git status -s)
if [ -z "$areChanges"  ]; then {
echo "There are no changes in the mak_common library"
echo "sv.sh ends its work ------------------------------------------------------"
exit
} 
fi

version=$1
toGit=$2
if [ -z "$version" ]; then {
echo "There are no version was passed to here"
echo "v.sh ended its work ------------------------------------------------------"
}
fi


#version="+++da21c61--*main--201225_1532+++"
#fix=+++
#version=$(echo "$version" | sed -e "s/^$fix//" -e "s/$fix$//")
echo version=$version






{
sed -i "s/---.*---/---$version---/" ksess/rules.txt
goOutOnError "sed ksess/rules.txt"

sed -i "s/---.*---/---$version---/" ksess/api.txt
goOutOnError "sed ksess/api.txt"

sed -i "s/---.*---/---$version---/" ksess/version.go
goOutOnError "sed ksess/version.go"

#210105 05:27
sed -i "s/---.*---/---$version---/" kconfig/config.go
goOutOnError "sed ksess/version.go"
}


if [ -z "$toGit" ]; then {
echo "There are no demand to git"
echo "v.sh ended its work ------------------------------------------------------"
exit
}
fi





git add .
goOutOnError "git add ."

git commit -m "$version:$toGit"
goOutOnError "git commit"

git push
goOutOnError "git push"

echo sv.sh: successfully ended with changes pushed------------------------------------------------------








