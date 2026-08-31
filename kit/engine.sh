#!/bin/bash
# Shim de compatibilité — le moteur vit maintenant dans le binaire : `cgo kit`.
# Ce script transmet les arguments tels quels (mêmes actions, mêmes codes de
# sortie). Retiré à v1.2 : utilisez `cgo kit <action>` directement.
exec cgo kit "$@"
