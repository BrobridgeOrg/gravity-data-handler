#!/bin/bash

[ "$#" -eq 0 ] || {

	echo $@ > ./rules/rules.json

}

/gravity-data-handler
