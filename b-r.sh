#!/bin/bash

# It is the template for writing a script that builds a golang ecxec file with setting version number (see /home/mich412/go/src/mak_common/mversion/ecxec.go)
# It takas one param ($1) It $1=="toGit" (see ~/go/src/mak_common/sv.sh)




# Preparing for calling ~/go/src/mak_common/sv.sh
mak_common_dir=~/go/src/mak_common
working_dir=$PWD

curr_dir_name=$(basename working_dir)




{
git_commit_info=$(git log --pretty=format:"%h" -n 1)
compiltime=$(date +%y%m%d_%H%M)
}

if [ $? != 0 ]; then 
echo getting the build time or the git infomation failed
exit
fi



#git branch --contains $git_commit_1 > rels_current_branch.txt

#sed 's/ //g' rels_current_branch.txt > rels_current_branch_without.txt

#curr_br=$(cat rels_current_branch_without.txt)

curr_branch=$(git branch --contains $git_commit_1 | sed 's/ //g')



echo ------------------------------------
echo commit = $git_commit_info
echo branch = $curr_br
echo ================================
git_commit_info=($git_commit_info--$curr_br--$compiltime)
echo ================================
echo git_commit_info = $git_commit_info
echo ------------------------------------




cd $mak_common_dir
~/go/src/mak_common/sv.sh "$curr_dir_name:$git_commit_info" $1
cd $working_dir



#go build -ldflags "-X main.git_commit_1=$git_commit_1"
go build -ldflags "-X main.git_commit_1=$git_commit_1"
if [ $? != 0 ]; then 
echo golang building failed
exit
fi

sed -i "s/nv.*env/nv $git_commit_info env/" version.go 


./$curr_dir_name




