#!/bin/sh

if [ -z "$1" ]; then
  echo "Listing all AWS services"
  cat index.json | grep offerCode | sort
else
  echo "Matches for for $1:"
  grep -i $1 index.json | grep offerCode
fi

