#!/bin/bash

[ "$#" -eq 0 ] || {

	echo $@ > ./rules/rules.json

}

exec /gravity-data-handler
