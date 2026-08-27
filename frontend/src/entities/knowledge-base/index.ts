export { default as KnowledgeBaseRepository } from './api/knowledgeBaseRepository';
export { flattenKbTree, collectKbAncestorIds } from './lib/tree';
export type { KbTreeRow } from './lib/tree';
export type {
  KbWorkspace,
  KbSpace,
  KbPage,
  KbPageTreeNode,
  KbPageTree,
  KbPageDoc,
} from './model/types';
