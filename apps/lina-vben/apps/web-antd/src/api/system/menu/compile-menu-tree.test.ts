import { describe, expect, it } from 'vitest';

import { compileMenuTree } from './compile-menu-tree';

describe('compileMenuTree', () => {
  it('builds a parent/child tree from a flat assignable menu list', () => {
    const tree = compileMenuTree([
      { id: 2, parentId: 1, label: 'Users', type: 'M' },
      { id: 1, parentId: 0, label: 'System', type: 'D' },
      { id: 3, parentId: 2, label: 'Create', type: 'B' },
    ]);

    expect(tree).toHaveLength(1);
    expect(tree[0]?.label).toBe('System');
    expect(tree[0]?.children?.map((item) => item.label)).toEqual(['Users']);
    expect(tree[0]?.children?.[0]?.children?.map((item) => item.label)).toEqual([
      'Create',
    ]);
  });
});
