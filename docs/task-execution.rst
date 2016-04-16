Task execution
==============

Now that you have a global idea of how the pythia-core framework is architectured, let us examine how to create a new task and then submit a job to execute it with the framework.



Task filesystem
---------------

As previously mentioned, a task is just a set of files in a given directory. This latter is then `compressed` into an ``.sfs`` file using the `SquashFS filesystem
<http://squashfs.sourceforge.net/>`_. Two elements must be kept in mind when creating a task filesystem:

* It must contain a ``control`` file with the sequence of executables to launch. 
* The task filesystem will be mounted in the ``/task`` directory inside the virtual machine.

The `Hello World` example presented in the :doc:`presentation section</presentation>` is composed of two files is the ``hello-world`` directory:

.. code-block:: none

   hello-world/
      control
      hello.sh

To build a `SquashFS filesystem` for this directory, you just have to use ``mksquashfs``. The following command will create a ``hello-world.sfs`` file with the content of the ``hello-world`` directory:

.. code-block:: none

   > mksquashfs hello-world hello-world.sfs -all-root -comp lzo -noappend

Note that you can extract the files contained in a SquashFS file with the ``unsquashfs`` command:

.. code-block:: none

   > unsquashfs -d hello-world hello-world.sfs



Task execution
--------------

As previously mentioned, a task can be easily executed by the pythia-core framework with the ``execute`` subcommand. For that to be possible, an additional ``.task`` file containing the configuration of the task must be created. This file refers in particular to the ``.sfs`` file with the SquashFS filesystem of the task. The input to provide for the execution of the task is stored in a text file, which can be empty.

The task can be executed as a job in the pythia-core framework with the following command, assuming that the ``pythia`` executable is in your ``PATH`` and that you are in the directory containing the three files ``hello-world.sfs``, ``hello-world.task`` and ``input.txt``:

.. code-block:: none

   > pythia execute -input="input.txt" -task="hello-world.task"
   Warning: unable to read configuration file: open config.json: no such file or directory
   Status: success
   Output: Hello world!

As you can observe, the ``execute`` subcommand produces two pieces of information: the status of the execution and the output produced by the task. We can also notice a warning about a missing configuration file, described hereafter.


Configuration
`````````````

The options of the ``pythia`` executables can be specified directly through the command line or thanks to a ``config.json`` file. This file is not mandatory but can be useful when :doc:`setting up the pythia-core framework</setup>`. The configuration is a JSON object composed of a global section and one specific section for each available subcommand. To know what are the possible configuration options, please refer to the :doc:`usage page</usage>` of this documentation.

Hereafter is an example of a ``config.json`` file:

.. code-block:: json

   {
     "global": {
       "queue": "127.0.0.1:9000"
     },
     "components": [
       {
         "component": "queue",
         "capacity": "500"
       },
       {
         "component": "pool"
       }
     ]
   }


Execution status
````````````````

There are `seven different status` for the execution of a task, summarised in the table hereafter. Depending on the status, the output takes different values.

.. table::

   +--------------+----------------------------------------------+---------------+
   | Status       | Description                                  | Output        |
   +==============+==============================================+===============+
   | ``success``  | Finished                                     | stdout        |
   +--------------+----------------------------------------------+---------------+
   | ``timeout``  | Timed out                                    | stdout so far |
   +--------------+----------------------------------------------+---------------+
   | ``overflow`` | stdout too big                               | capped stdout |
   +--------------+----------------------------------------------+---------------+
   | ``abort``    | Aborted by abort message                     | --            |
   +--------------+----------------------------------------------+---------------+
   | ``crash``    | Sandbox crashed                              | stdout        |
   +--------------+----------------------------------------------+---------------+
   | ``error``    | Error (maybe temporary)                      | error message |
   +--------------+----------------------------------------------+---------------+
   | ``fatal``    | Unrecoverable error (e.g. misformatted task) | error message |
   +--------------+----------------------------------------------+---------------+
   