Presentation
============

The pythia-core framework makes it possible to execute code in a safe environment. It can be used to grade codes written written by non-trusted people, such as students in a learning environment, for example. The framework can execute jobs provided with optional inputs and that will produce outputs upon completion.



Queue and pools
---------------

The pythia-core framework is built on two mains components, namely one queue and several pools. As shown on Figure 1, the `queue` is the entry point of the framework which receives the jobs to be executed from the outside. It then dispatches the jobs to pools (of execution) and waits for the result to send it back to the job's submitter. A `pool` launches a new virtual machine for each job it is asked to execute, so that to execute it in a safe and controlled environment.

.. figure:: _static/pythia-core-general-view.png
   :align: center
   :scale: 40 %
   :alt: map to buried treasure

   Figure 1: This is the caption of the figure (a simple paragraph).

The queue is the main component and is waiting for incoming connections. It means that the other components, that is, the pools and the frontends, have to first connect to the queue before being able to offer their services.



Environment and task filesystems
--------------------------------

`Jobs` that are executed by the virtual machines launched by the pools are composed of several elements as shown on Figure 2.

.. figure:: _static/pythia-core-job-structure.png
   :align: center
   :scale: 40 %
   :alt: map to buried treasure
   
   Caption of the figure