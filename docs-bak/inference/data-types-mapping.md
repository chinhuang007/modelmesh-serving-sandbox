---
title: Data Types Mapping
---

### PyTorch, TensorFlow and ONNX

Refer to Triton documentation on [Data types Mapping](https://github.com/triton-inference-server/server/blob/main/docs/model_configuration.md#datatypes)

### LightGBM, Sklearn and XGBoost

Seldon MLServer is fully compliant with KFServing's V2 Dataplane spec. Refer here for more information on [KFServing V2 Tensor data types](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/required_api.md#tensor-data-types)

### Spark, PMML and MLeap

The following table shows the tensor data types supported by MSP MLServer and how to map the data types of models supported by MSP MLServer to tensor data types.

| Spark         | PMML      | MLeap   | Tensor Types                |
| ------------- | --------- | ------- | --------------------------- |
| ByteType      |           | Byte    | INT8                        |
| BinaryType    |           |         | BYTES                       |
| ShortType     |           | Short   | INT16                       |
| IntegerType   | Integer   | Integer | INT32                       |
| LongType      |           | Long    | INT64                       |
| FloatType     |           | Float   | FP32                        |
| DoubleType    | Real      | Double  | FP64                        |
| StringType    | String    | String  | STRING                      |
| BooleanType   |           | Boolean | BOOL                        |
| DateType      | Date      |         | DATE                        |
| TimestampType | Timestamp |         | TIMESTAMP                   |
| ArrayType     |           | Array   | Any of above primitive type |
| DecimalType   |           |         | FP64                        |
| VectorType    |           | Tensor  | VECTOR                      |
