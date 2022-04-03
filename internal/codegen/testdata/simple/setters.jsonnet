local types = import 'types.libsonnet';
local bytesString = std.base64('hello');

types.testdata.simple.TopMessage.
  withBoolField(true).
  withBytesField(bytesString).
  withDoubleField(12.2).
  withEnumField(types.testdata.simple.TopLevelEnum.FIRST).
  withFixed32Field(7).
  withFixed64Field(8).
  //withFloatField(11.1).
  withInner1(
  types.testdata.simple.TopMessage.InnerMessage1.
    withNumbers(['ONE', 'TWO'])
).
  withInner2(types.testdata.simple.TopMessage.InnerMessage2.withMsgs(
  {
    m1: { numbers: ['TWO'] },
  }
)).
  withInt32Field(1).
  withInt64Field(2).
  withSfixed32Field(9).
  withSfixed64Field(10).
  withSint32Field(3).
  withSint64Field(4).
  withStrField('string').
  withUint32Field(5).
  withUint64Field(6).
  _validate()
