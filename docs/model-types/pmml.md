---
title: PMML
description: PMML model
---

## Format

> Predictive Model Markup Language (PMML) is an XML-based standard established by the Data Mining Group (DMG) for defining statistical and data-mining models. PMML models can be shared between PMML-compliant platforms and across organizations so that business analysts and developers are unified in designing, analyzing, and implementing PMML-based assets and services.

([source](http://dmg.org/pmml/pmml-v4-4.html))

For more information about the background and applications of PMML, see the DMG ([PMML specification](http://dmg.org/pmml/v4-4/GeneralStructure.html)).

Model-Mesh Serving supports PMML model file in XML format.

## Storage Layout

storage path can point directly to XML file

```
<storage-path/model-folder.xml>
```

### Example

**Storage Layout**

```
s3://wml-serving-examples/
└── pmml-models
    └── example-model.xml
```

**Predictor**

```
apiVersion: wmlserving.ai.ibm.com/v1
kind: Predictor
metadata:
  name: pmml-example
spec:
  modelType:
    name: pmml
  path: pmml-models/example-model.xml
  storage:
    s3:
      secretKey: modelStorage
      bucket: wml-serving-examples
```
