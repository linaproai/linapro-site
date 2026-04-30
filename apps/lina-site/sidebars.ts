import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  // 需要按照版本区分的文档：开发手册
  mainSidebar: [
    'docs/overview',
    {
      type: 'category',
      label: 'Core Concepts',
      collapsed: false,
      items: [
        'docs/core-concepts/layers',
        'docs/core-concepts/built-in-capabilities',
      ],
    },
    {
      type: 'category',
      label: 'Architecture',
      collapsed: false,
      items: ['docs/architecture/runtime-architecture'],
    },
    {
      type: 'category',
      label: 'Plugin Development',
      collapsed: false,
      items: ['docs/plugin-development/overview'],
    },
    {
      type: 'category',
      label: 'Testing and Deployment',
      collapsed: false,
      items: ['docs/testing-deployment/overview'],
    },
  ],

  // 快速开始：面向首次体验的教程与示例
  quickSidebar: [
    'quick/preface/preface',
    {
      type: 'category',
      label: 'First Experience',
      collapsed: false,
      items: [
        'quick/quickstart',
        'quick/workspace-tour',
        'quick/first-plugin',
      ],
    },
  ],

  // 开源社区：交流、贡献与支持入口
  communitySidebar: [
    'community/community',
    {
      type: 'category',
      label: 'Community Channels',
      collapsed: false,
      items: ['community/channels'],
    },
    {
      type: 'category',
      label: 'Contribution',
      collapsed: false,
      items: ['community/contribution'],
    },
    {
      type: 'category',
      label: 'Support',
      collapsed: false,
      items: ['community/support/donate'],
    },
  ],
};

export default sidebars;
