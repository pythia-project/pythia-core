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




Queue
-----




Pool
----




Server
------



