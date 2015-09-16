#!/bin/sh

cd ~/go/src/jtwsm.net/code/
git pull origin master
cd ./blog/
go install
cd ~/jtwsm.net/apps/
git pull origin master
killall blog
cd ./blog/
cp ~/go/bin/blog ./
nohup ./blog -http=:8080 &
