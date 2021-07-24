#!/bin/bash

scp build $1:/opt/debt_bot/build

ssh $1 './restart_debt.sh'
