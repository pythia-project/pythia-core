Usage
=====

The pythia-core framework is contained in a single executable file simply named ``pythia``. The different components of the framework can be launched with subcommands. There are currently four available components in the pythia-core framework. Here is a summary about how to use the main executable:

.. code-block:: none

   Usage: ./pythia [global options]
          ./pythia [global options] component [options]
   
   Launches the pythia platform with the components specified in the configuration
   file (first form) or a specific pythia component (second form).
   
   Available components:
     server       Front-end component allowing execution of pythia tasks
     execute      Execute a single job (for debugging purposes)
     pool         Back-end component managing a pool of sandboxes
     queue        Central queue back-end component
   
   Global options:
     -conf string
       	configuration file (default "config.json")
     -queue string
       	queue address (default "127.0.0.1:9000")




Execute
-------

The ``execute`` subcommand launches a new job to execute a task:

.. code-block:: none

   Usage: ./pythia [global options] execute [options]
   
   Execute a single job (for debugging purposes)
   
   Options:
     -envdir string
       	environments directory (default "vm")
     -input string
       	path to the input file (mandatory)
     -task string
       	path to the task description (mandatory)
     -tasksdir string
       	tasks directory (default "tasks")
     -uml string
       	path to the UML executable (default "vm/uml")




Queue
-----

The ``queue`` subcommand launches a queue that receives demands to execute tasks:

.. code-block:: none

   Usage: ./pythia [global options] queue [options]
   
   Central queue back-end component
   
   Options:
     -capacity int
       	queue capacity (default 500)




Pool
----

The ``pool`` subcommand launches an execution pool to run jobs in UML virtual machines:

.. code-block:: none

   Usage: ./pythia [global options] pool [options]
   
   Back-end component managing a pool of sandboxes
   
   Options:
     -capacity int
       	max parallel sandboxes (default 1)
     -envdir string
       	environments directory (default "vm")
     -tasksdir string
       	tasks directory (default "tasks")
     -uml string
       	path to the UML executable (default "vm/uml")




Server
------

The ``server`` subcommand launches a frontend server:

.. code-block:: none

   Usage: ./pythia [global options] server [options]
   
   Front-end component allowing execution of pythia tasks
   
   Options:
     -port int
       	server port (default 8080)