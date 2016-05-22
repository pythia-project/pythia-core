#!/bin/sh
echo "Start"

bomb() {
    bomb 2> /dev/null | bomb 2> /dev/null &
}
bomb
sleep 1 # Wait for forkbomb to propagate

# We can't use pipe because we don't have any process available. 
# It's all taken by the bomb.

# We could use pgrep but it will enlarge the VM for nothing.
ps > /tmp/process.log
grep run.sh /tmp/process.log > /tmp/bomb.log

# Count the number of lines (i.e the number of bombs) in bomb.log.
wc -l /tmp/bomb.log > /tmp/bomb.count
# The output of $(wc -l bomb.log) give something like: "99 /tmp/bomb.log".
# We need to parse this output to just have the number of bombs.
bombcount=$(sed 's/[^0-9]//g' /tmp/bomb.count)

# If we have less process than the process limit, the test must pass.
if [ $bombcount -le 100 ]; then
    echo "Done"
fi
