export { default as KnowledgeBaseRepository } from './api/knowledgeBaseRepository';
export { flattenKbTree, collectKbAncestorIds, replaceKbPageInTree } from './lib/tree';
export type { KbTreeRow, KbHiddenRow, KbTreeEntry } from './lib/tree';
export type {
  KbWorkspace,
  KbSpace,
  KbPage,
  KbPageTreeNode,
  KbPageTree,
  KbPageDoc,
} from './model/types';
