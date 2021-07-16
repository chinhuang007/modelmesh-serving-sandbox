---
title: Extended KFServing V2 gRPC API
---

This section is currently applicable only to those model types handled by the MSP MLServer: Spark, MLeap, PMML.

### 1. Extended KFServing V2 gRPC protocol to support new datatypes

#### a) STRING DataType

STRING datatype to represent strings and content must be provided in "bytesContents"

#### Example

Tensor with **STRING** datatype

```
{
     "name": "my_string_col",
     "datatype": "STRING",
     "shape": [ 1 ],
     "contents": {
         "bytesContents": [ b"string data in bytes format" ]
     }
}
```

#### b) DATE DataType

DATE datatype to represent date and content must be provided in "bytesContents". The supported format for date is "YYYY-MM-DD"

#### Example

Tensor with **DATE** datatype

```
{
     "name": "my_date_col",
     "datatype": "DATE",
     "shape": [ 1 ],
     "contents": {
         "bytesContents": [ b"YYYY-MM-DD" ]
     }
}
```

#### c) TIMESTAMP DataType

TIMESTAMP datatype to represent the timestamp and content must be provided in "bytesContents". The supported format for the timestamp is "dd/mm/yyyy HH:MM:SS"

#### Example

Tensor with **TIMESTAMP** datatype

```
{
     "name": "my_timestamp_col",
     "datatype": "TIMESTAMP",
     "shape": [ 1 ],
     "contents": {
         "bytesContents": [ b"dd/mm/yyyy HH:MM:SS" ]
     }
}
```

#### d) VECTOR DataType

VECTOR datatype to represent the dense vectors in **Spark** and dense tensor in **MLeap**. Content must be provided in "fp64Contents"

To represent a VECTOR in tensor format, you must associate shape tensor along with the actual tensor where actual tensor includes the contents of dense vectors. [For more information on shape tensor](#5.-represent-complex-data-types-in-the-tensor:)

#### Example

Tensor with **VECTOR** datatype to represent a dense vectors with 2 rows and 1 field of variable size

**row 1**: [ 1.0, 2.0, 3.0 ]

**row 2**: [ 4.0, 5.0 ]

```
{
     "name": "my_vector_col",
     "datatype": "FP64",
     "shape": [ 2, 1, -1 ],
     "contents": {
        "fp64Contents": [ 1.0, 2.0, 3.0, 4.0, 5.0 ]
     }
},
{
     "name": "my_vector_col_shape"
     "datatype": "INT32",
     "shape": [ 1 ]
     "contents": {
        "int_contents": [ 3, 2 ]
     }
}
```

### 2. Represent decimal type in the tensor

Decimal type is represented using the tensor datatype "FP64". Content must be provided in "fp64Contents".

"precision" and "scale" must be provided in "parameters" section with type "string_param".

#### Example

Tensor with **FP64** type to represent decimal values <br/>

Here "precision" and "scale" are provided in "parameters" section

```
{
     "name": "my_decimal_col",
     "datatype": "FP64",
     "shape": [ 2 ],
     "parameters": {
        "precision": "10",
        "scale": "5"
     },
     "contents": {
         "fp64Contents": [ 10.12000, 20.00000 ]
     }
}
```

### 3. Encoding format

Supported encoding format to construct the "bytesContents" is **UTF-8**

#### Example

String content is encoded to UTF-8 format and provided in bytesContents as shown below:

```
{
     "name": "my_string_col",
     "datatype": "STRING",
     "shape": [ 1 ],
     "contents": {
         "bytesContents": [ <BYTE CONTENT in UTF-8 FORMAT> ]
     }
}
```

### 4. Fields to Tensor mapping

### a) Single field per tensor:

Field name must be provided in tensor name.

#### Example

Tensor with 4 rows and 1 field. Here tensor name "my_int_col" is field name

```
{
     "name": "my_int_col",
     "datatype": "INT32",
     "shape": [ 4 ],
     "contents": {
         "int_contents": [ 10, 19, 76, 4 ]
     }
}
```

### b) Multiple fields per tensor:

If consecutive fields are of same datatype, it can be grouped inside a single tensor. Field names must be provided in "parameters" of tensor with the name "features" and value containing the comma separated list of fields of type "string_param". Tensor name would be ignored.

See [Parameters](https://github.com/kubeflow/kfserving/blob/master/docs/predict-api/v2/required_api.md#parameters-1) for more information.

#### Examples

**1. Tensor with 1 row and 3 fields**

Consecutive field names "my_fp64_col1", "my_fp64_col2", "my_fp64_col3" are of "FP64" datatype. Field names are provided in "parameters" of tensor.

```
{
     "name": "FeatureCols1_3",
     "datatype": "FP64",
     "shape": [ 1, 3 ],
     "parameters": {
        "features": "my_fp64_col1,my_fp64_col2,my_fp64_col3"
     },
     "contents": {
         "fp64Contents": [ 1.01, 1.02, 1.03 ]
     }
}
```

**2. Tensor with 3 rows and 2 fields**

Consecutive field names "my_int_col1" and "my_int_col2" are of "INT32" datatype. Field names are provided in "parameters" section of tensor.

**row 1**: &nbsp;=> &nbsp;[ 10, 11] &nbsp;=> &nbsp;[ my_int_col1, my_int_col2 ]

**row 2**: &nbsp;=> &nbsp;[ 20, 21] &nbsp;=> &nbsp;[ my_int_col1, my_int_col2 ]

**row 3**: &nbsp;=> &nbsp;[ 30, 31] &nbsp;=> &nbsp;[ my_int_col1, my_int_col2 ]

```
{
     "name": "FeatureCols1_2",
     "datatype": "INT32",
     "shape": [ 3, 2 ],
     "parameters": {
        "features": "my_int_col1,my_int_col2"
     },
     "contents": {
         "int_contents": [ 10, 11, 20, 21, 30, 31 ]
     }
}
```

### 5. Represent complex data types in the tensor:

Complex data types like Array and Vector can have fields with variable size dimensions. To represent the complex data types in tensor representation, you must associate shape tensor along with the actual tensor where actual tensor includes the contents of complex data types.
Shape tensor must include dimensions of actual tensor with datatype "INT32".

Shape tensor is identified based on tensor name. If actual tensor name is "T" then shape tensor name must be "T_shape" where "T" can be of any name.

Actual tensor "T" of complex data type must have shape: [ R, F, -1 ] where "R" represents the number of rows and "F" represents the number of fields with variable size. Shape tensor "T_shape" must have shape: [ 1 ] for now.

#### Example 1: Tensors representing Array<FP64\> with 3 rows and 1 field of variable size

The actual tensor "my_array_col" representing the datatype **ARRAY<FP64\>** with 3 rows and 1 field of variable size

Shape tensor "my_array_col_shape" is having dimension of actual tensor with datatype **INT32**.

**row 1:** &nbsp;[ 1.0, 2.0, 3.0 ] &nbsp;=> &nbsp;[ "my_array_col" ]

**row 2:** &nbsp;[ 4.0, 5.0 ] &nbsp; &nbsp; &nbsp; &nbsp; => &nbsp; [ "my_array_col" ]

**row 3:** &nbsp;[ 6.0 ] &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;=> &nbsp; [ "my_array_col" ]

```
{
     "name": "my_array_col",
     "datatype": "FP64",
     "shape": [ 3, 1, -1 ],
     "contents": {
        "fp64Contents": [ 1.0, 2.0, 3.0, 4.0, 5.0, 6.0 ]
     }
},
{
     "name": "my_array_col_shape"
     "datatype": "INT32",
     "shape": [ 1 ]
     "contents": {
        "int_contents": [ 3, 2, 1 ]
     }
}
```

#### Example 2: Tensors representing Array<STRING\> with 2 rows and 2 fields of variable size

The actual tensor "FeatureCols1_2" representing the datatype **ARRAY<STRING\>** with 2 rows and 2 field of variable size.

Shape tensor "FeatureCols1_2_shape" is having dimension of actual tensor with datatype **INT32**.

**row 1:** <br />
&nbsp; &nbsp; &nbsp; field 1 : &nbsp;[ "string1", "string2" ] <br />
&nbsp; &nbsp; &nbsp; field 2 : &nbsp;[ "string3" ] &nbsp;
<br />
<br />
**row 2:** <br />
&nbsp; &nbsp; &nbsp; field 1 : &nbsp; [ "string4", "string5", "string6" ] <br />
&nbsp; &nbsp; &nbsp; field 2 : &nbsp; [ "string7", "string8" ]

```
{
     "name": "FeatureCols1_2",
     "datatype": "STRING",
     "shape": [ 2, 2, -1 ],
     "parameters": {
        "features": "my_array_col1,my_array_col2"
     },
     "contents": {
        "fp64Contents": [ "string1", "string2", "string3", "string4", "string5", "string6", "string7", "string8" ]
     }
},
{
     "name": "FeatureCols1_2_shape"
     "datatype": "INT32",
     "shape": [ 1 ]
     "contents": {
        "int_contents": [ 2, 1, 3, 2 ]
     }
}
```

### 6. Model schema

- Schema is optional for now.

- Schema is built from KFServing V2 GRPC input

### Example

Model infer request representing the two input tensors where Tensor 1 is single field per tensor and Tensor 2 is multiple fields per tensor.

Model schema is built using the input field names and datatypes.<br/>
**Field names:** my_field_col1, my_field_col2, my_field_col3 <br/>
**DataTypes :** FP64, STRING, STRING

```
{
     "name": "my_field_col1",
     "datatype": "FP64",
     "shape": [ 1 ],
     "contents": {
         "fp64Contents": [ 0.10 ]
     }
},
{
     "name": "featureCols2_3",
     "datatype": "STRING",
     "shape": [ 1, 2 ],
     "parameters": {
        "features": "my_field_col2,my_field_col3"
     },
     "contents": {
         "bytesContents": [ b"first content", b"second content" ]
     }
}
```
