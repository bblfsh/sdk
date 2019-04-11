#!/usr/bin/env bash
# This implements a trivial native driver that returns its input unchanged.
# TODO: Implement a parser for the target language.
while read -r req
do
  resp='{"status":"ok", "ast": '
  resp+=$req
  resp+='}'
  echo $resp
done
exit 1
