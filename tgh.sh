#!/bin/bash
echo it is tgh.sh -- saving to git hub

#exit

git_commit_1=no_git???
compiltime=No_compilation_time

date +%y%m%d_%H%M > compilation_time.txt
if [ $? != 0 ]; then 
echo getting the saving time failed
exit
fi


git log --pretty=format:"%h" -n 1 > git_commit_1.txt
if [ $? != 0 ]; then 
echo getting the git commit information failed
exit
fi

git_commit_1=$(cat git_commit_1.txt)
compiltime=$(cat compilation_time.txt)


git branch --contains $git_commit_1 > rels_current_branch.txt
if [ $? != 0 ]; then 
echo getting the git branch information failed
exit
fi



sed 's/ //g' rels_current_branch.txt > rels_current_branch_without.txt
curr_br=$(cat rels_current_branch_without.txt)


echo ------------------------------------
echo commit = $git_commit_1
echo branch = $curr_br
echo ================================
git_commit_1=($git_commit_1--$curr_br--$compiltime)
echo ================================
echo $git_commit_1
echo ------------------------------------

{
sed -i "s/---.*---/---$git_commit_1---/" ksess/rules.txt
}


rm git_commit_1.txt
rm compilation_time.txt
rm rels_current_branch.txt
rm rels_current_branch_without.txt

echo $1
#exit
#201223 05:15 does not roll
git add .
echo passed git add . 

git commit -m "$1"
echo git commit -m "$1"


git push








