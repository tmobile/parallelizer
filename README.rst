============
Parallelizer
============

.. image:: https://img.shields.io/github/tag/klmitch/parallelizer.svg
    :target: https://github.com/klmitch/parallelizer/tags
.. image:: https://img.shields.io/hexpm/l/plug.svg
    :target: https://github.com/klmitch/parallelizer/blob/master/LICENSE
.. image:: https://travis-ci.org/klmitch/parallelizer.svg?branch=master
    :target: https://travis-ci.org/klmitch/parallelizer
.. image:: https://coveralls.io/repos/github/klmitch/parallelizer/badge.svg?branch=master
    :target: https://coveralls.io/github/klmitch/parallelizer?branch=master
.. image:: https://godoc.org/github.com/klmitch/parallelizer?status.svg
    :target: http://godoc.org/github.com/klmitch/parallelizer
.. image:: https://img.shields.io/github/issues/klmitch/parallelizer.svg
    :target: https://github.com/klmitch/parallelizer/issues
.. image:: https://img.shields.io/github/issues-pr/klmitch/parallelizer.svg
    :target: https://github.com/klmitch/parallelizer/pulls
.. image:: https://goreportcard.com/badge/github.com/klmitch/parallelizer
    :target: https://goreportcard.com/report/github.com/klmitch/parallelizer

This repository contains Parallelizer.  Parallelizer is a library for
enabling the addition of controlled parallelization utilizing a pool
of worker goroutines in a simple manner.  This is not intended as an
external job queue, where outside programs may submit jobs, although
it could easily be used to implement such a tool.

Testing
=======

This repository is a standard go repository, and so may be tested and
built in the standard go ways.  However, the repository also contains
a ``Makefile`` to aid in repeatable testing and reformatting;
developers that wish to contribute to Parallelizer may find it useful
to utilize ``make`` to ensure that their code conforms to the
standards enforced by Travis CI.  The following is a run-down of the
available ``make`` targets.

``make format-test``
--------------------

This target is called by Travis to ensure that the formatting conforms
to that recommended by the standard go tools ``goimports`` and
``gofmt``.  Most developers should prefer the ``make format`` target,
which is automatically run by ``make test`` or ``make cover``, and
will rewrite non-conforming files.  Note that ``goimports`` is a
third-party package; it may be installed using::

    % go get -u -v golang.org/x/tools/cmd/goimports

``make format``
---------------

This target may be called by developers to ensure that the source code
conforms to the recommended style.  It runs ``goimports`` and
``gofmt`` to this end.  Most developers will prefer to use ``make
test`` or ``make cover``, which automatically invoke ``make format``.
Note that ``goimports`` is a third-party package; it may be installed
using::

    % go get -u -v golang.org/x/tools/cmd/goimports

``make lint``
-------------

This target may be called to run a lint check.  This tests for such
things as the presence of documentation comments on exported functions
and types, etc.  To this end, this target runs ``golint`` in enforcing
mode.  Most developers will prefer to use ``make test`` or ``make
cover``, which automatically invoke ``make lint``.  Note that
``golint`` is a third-party package; it may be installed using::

    % go get -u -v golang.org/x/lint/golint

``make vet``
------------

This target may be called to run a "vet" check.  This vets the source
code, looking for common problems prior to attempting to compile it.
Most developers will prefer to use ``make test`` or ``make cover``,
which automatically invoke ``make vet``.

``make test-only``
------------------

This target may be called to run only the unit tests.  A coverage
profile will be output to ``coverage.out``, but none of the other
tests, such as ``make vet``, will be invoked.  Most developers will
prefer to use ``make test`` or ``make cover``, which automatically
invoke ``make test-only``, among other targets.

``make test``
-------------

This target may be called to run all the tests.  It ensures that
``make format``, ``make lint``, ``make vet``, and ``make test-only``
are all called, in that order.

``make cover``
--------------

This target may be called to run ``make test``, but will additionally
generate an HTML file named ``coverage.html`` which will report on the
coverage of the source code by the test suite.

``make clean``
--------------

This target may be called to remove the temporary files
``coverage.out`` and ``coverage.html``, as well as any future
temporary files that are added in the testing process.

Contributing
============

Contributions are welcome!  Please ensure that all tests described
above pass prior to proposing pull requests; pull requests that do not
pass the test suite unfortunately cannot be merged.  Also, please
ensure adequate test coverage of additional code and branches of
existing code; the ideal target is 100% coverage, to ensure adequate
confidence in the function of Parallelizer.
