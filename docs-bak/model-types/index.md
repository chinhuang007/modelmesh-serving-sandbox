---
title: Supported Model Formats
---

By leveraging existing third-party model servers, we support a number of standard ML model formats out-of-the box, with more to follow. Currently supported model types:

- [LightGBM](model-types/lightgbm)
- [MLeap](model-types/mleap)
- [ONNX](model-types/onnx)
- [PMML](model-types/pmml)
- [PyTorch ScriptModule](model-types/pytorch)
- [scikit-learn](model-types/sklearn)
- [Spark](model-types/spark)
- [TensorFlow](model-types/tensorflow)
- [XGBoost](model-types/xgboost)

| Model Type | Framework    | Versions        | Supported via ServingRuntime | Since      |
| ---------- | ------------ | --------------- | ---------------------------- | ---------- |
| lightgbm   | LightGBM     | 3.2.1           | MLServer (python)            | 0.5.0      |
| mleap      | MLeap        | 0.17            | MSP MLServer (Scala)         | 0.2.0\(\*) |
| onnx       | ONNX         | 1.5.3           | Triton (C++)                 | 0.4.1      |
| pmml       | PMML         | 3.x, 4.x        | MSP MLServer (Scala)         | 0.5.0      |
| pytorch    | PyTorch      | 1.8.0a0+1606899 | Triton (C++)                 | 0.3.0      |
| sklearn    | scikit-learn | 0.23.1          | MLServer (python)            | 0.3.0      |
| spark      | Spark MLlib  | 3.1.1           | MSP MLServer (Scala)         | 0.5.0      |
| tensorflow | TensorFlow   | 1.15.4, 2.3.1   | Triton (C++)                 | 0.2.0      |
| xgboost    | XGBoost      | 1.1.1           | MLServer (python)            | 0.4.1      |
| \*         | Custom       |                 | [Custom](runtimes) (any)     | 0.2.0      |

(\*) MLeap models were handled by a separate mleap-serving runtime pre-0.5.0
