export interface MenuTreeNode {
  id: number;
  parentId: number;
  label: string;
  type: string;
  icon?: string;
  children?: MenuTreeNode[];
}

/** Compile a flat assignable menu list into a parent/child tree. */
export function compileMenuTree(nodes: MenuTreeNode[]): MenuTreeNode[] {
  const map = new Map<number, MenuTreeNode>();
  for (const node of nodes) {
    map.set(node.id, { ...node, children: [] });
  }
  const roots: MenuTreeNode[] = [];
  for (const node of map.values()) {
    const parent = map.get(node.parentId);
    if (parent && parent !== node) {
      parent.children = parent.children ?? [];
      parent.children.push(node);
      continue;
    }
    roots.push(node);
  }
  return roots;
}
