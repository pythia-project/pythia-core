Task compilation and execution
==============================

Now that you have a global idea of how the pythia-core framework is architectured, let us examine how to create and compile a new task and submit a job to execute it.



Task compilation
----------------

As previously mentioned, a task is just a set of files in a given directory. This latter is then compressed into an ``.sfs`` file using the SquashFS filesystem. Two elements must be kept in mind when creating a task filesystem :

* It must contain a ``control`` file with the sequence of executables to laucnh. 
* The task filesystem will be mounted in the ``/task`` directory inside the virtual machine.



Task execution
--------------



Output specification
--------------------