# Pythia-core [![Build Status](https://travis-ci.org/pythia-project/pythia-core.svg?branch=master)](https://travis-ci.org/pythia-project/pythia-core) [![Documentation Status](https://readthedocs.org/projects/pythia-core/badge/?version=latest)](http://pythia-core.readthedocs.org/en/latest/?badge=latest) [![Join the chat at https://gitter.im/pythia-project/pythia](https://badges.gitter.im/pythia-project/pythia.svg)](https://gitter.im/pythia-project/pythia?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Pythia-core is the backbone of the Pythia framework. It manages a pool of UML virtual machines and is in charge of the safe execution of low-level jobs. Pythia-core is written in [Go](https://golang.org) and can be easily distributed on several machines or in the cloud.

## Quick Install

Since the pythia-core framework uses UML-based virtual machines, it can only be run on Linux.

Start by installing required dependencies:

- Make (4.0 or later)
- Go (1.2.1 or later)
- SquashFS tools (``squashfs-tools``)
- Embedded GNU C Library (``libc6-dev-i386``)

Then, clone the Git repository, and launch the installation:

    > git clone --recursive https://github.com/pythia-project/pythia-core.git
    > cd pythia-core
    > make

Once successfully installed, you can try to execute a simple task:

    > cd out
    > touch input.txt
    > ./pythia execute -input="input.txt" -task="tasks/hello-world.task"

and you will see, among others, ``Hello world!`` printed in your terminal.

## Use with Docker
Docker allow the pythia-core framework to run on MacOS or Windows installation.

Start by cloning the git repository and build the docker image:

    > git clone --recursive https://github.com/pythia-project/pythia-core.git
    > cd pythia-core
    > docker build -t pythia-core .

Once the image is successfully built, you can now start the image:

    > docker run -dit -p 8080:8080 --security-opt seccomp:unconfined --privileged pythia-core
    > docker exec -it --privileged CONTAINER_ID bash
    > mount /dev/shm
    > cd out && touch input.txt
    > ./pythia execute -input="input.txt" -tasks="tasks/hello-world.task"

You can obtain the container id using docker ps.
You should see among others, ``Hello world!`` printed in your terminal.

## Contributors

- Sébastien Combéfis
- Guillaume de Moffarts
- Vianney le Clément de Saint-Marcq
- Charles Vandevoorde
- Virginie Van den Schrieck
