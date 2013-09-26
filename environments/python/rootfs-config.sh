# Copyright 2013 The Pythia Authors.
# This file is part of Pythia.
#
# Pythia is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, version 3 of the License.
#
# Pythia is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with Pythia.  If not, see <http://www.gnu.org/licenses/>.

# Python 3.1
install_debs python3.1 python3-minimal python3.1-minimal

# Base libraries
install_debs libc6 libc-bin libgcc1

# Additional libraries
install_debs libexpat1 libssl0.9.8 zlib1g libbz2-1.0 libdb4.8 libncursesw5 \
             libreadline6 libncurses5 readline-common libsqlite3-0 mime-support
