Framework setup
===============

We now know how to create a new task and to execute it directly with the pythia-core framework. In this section we examine how to setup the pythia-core framework, that is, starting the queue and the pools and submitting tasks to the queue for execution.



Starting the queue
------------------

The main component of the pythia-core framework is the `queue`. Once started, the other components have to connect to the queue to register themselves and start using the services provided by the queue. The queue is characterised by two parameters:

* the address and port it is listening to (127.0.0.1:9000 by default);
* the maximum number of tasks waiting for execution (capacity of 500 by default).

Starting the queue with default parameters is as easy as:

.. code-block:: none

   > pythia queue
   Listening to 127.0.0.1:9000



Connecting pools
----------------

Once the queue is started, you have to start `execution pools` and connect them to the queue, so that this latter can use them to execute tasks. A pool is characterised by five parameters:

* the address and port of the queue to connect to (127.0.0.1:9000 by default);
* the maximum number of parallel running sandboxes (capacity of 1 by default);
* the directory where to find enviroments (``vm`` by default)
* the directory where to find tasks (``tasks`` by default)
* the path to the UML executable (``vm/uml`` by default)

Starting a new pool and connecting it to the queue is as easy as:

.. code-block:: none

   > pythia pool
   Connected to queue 127.0.0.1:9000

In the console of the queue, you can also notice that a new pool has been connected to the queue:

.. code-block:: none

   > pythia queue
   Listening to 127.0.0.1:9000
   Client 0: connected.
   Client 0: pool capacity 1

You can start as many pools as you want, as far as your machine is powerful enough to withstand the load. The queue will balance the tasks as equally as possible between all the pools that are connected to it.



Submitting a task with the server
---------------------------------

