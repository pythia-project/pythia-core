Pythia-core: safe code execution within UML virtual machines
============================================================

Pythia is a framework deployed as an online platform whose goal is to teach programming and algorithm design. The platform executes the code in a safe environment and its main advantage is to provide intelligent feedback to its users to suppor their learning. More details about the whole project can be found on the `official website of Pythia
<http://www.pythia-project.org/>`_.

Pythia-core is the backbone of the Pythia framework. It manages a pool of UML virtual machines and is in charge of the safe execution of low-level jobs. Pythia-core is written in `Go
<https://golang.org/>`_ and can be easily distributed on several machines or in the cloud.



Quick install
-------------

Since the pythia-core framework uses UML-based virtual machines, it can only be run on Linux.

Start by installing required dependencies:

* Make (4.0 or later)
* Go (1.2.1 or later)
* SquashFS tools (``squashfs-tools``)
* Embedded GNU C Library (``libc6-dev-i386``)

Then, clone the Git repository, and launch the installation:

.. code-block:: none

   > git clone --recursive https://github.com/pythia-project/pythia.git
   > cd pythia
   > make

Once successfully installed, you can try to execute a simple task:

.. code-block:: none

   > cd pythia/out
   > touch input.txt
   > ./pythia execute -input="input.txt" -task="tasks/hello-world.task"

and you will see ``Hello world!`` printed in your terminal.



Contents
--------

This documentation is split into two parts: the first one is targeter to users and the second one is for developers. In any case, we recommend you to first read the user's documentation to understand how to use and test the framework.


.. toctree::
   :maxdepth: 1
   :caption: User's Documentation
   
   presentation
   task-execution
   setup
   usage


.. toctree::
   :maxdepth: 1
   :caption: Developer's Documentation
   
   architecture
   commmsg