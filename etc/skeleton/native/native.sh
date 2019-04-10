#!/usr/bin/env bash
while read -r req
do
  resp='{"status":"ok", "ast": '
  resp+=$req
  resp+='}'
  echo $resp
done
exit 1
