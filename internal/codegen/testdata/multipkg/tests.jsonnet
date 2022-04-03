local basicTests = [
  {
    name: 'happy_path',
    summary: 'ensure that a basic multi package message can be created',
    code: |||
      local types = import 'types.libsonnet';
      types.testdata.multipkg.TopMessage.
        withLib(types.testdata.multipkg.inner.Lib.withName('lib')).
        withLib2(types.Lib2.withName('lib2')).
        _validate()
    |||,
    result: {
      lib: { name: 'lib' },
      lib2: { name: 'lib2' },
    },
  },
];

basicTests
