import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

import styles from './about.module.css';

const content = {
  en: {
    eyebrow: 'About LinaPro',
    title: 'An open-source AI engineering system designed to stay readable as it scales.',
    body:
      'LinaPro is built for teams that want AI-assisted delivery without giving up repository structure, review boundaries, or system clarity. The public site exists to explain that model in plain engineering terms.',
    primary: 'Open GitHub',
    secondary: 'Read architecture',
    signal: 'public operating notes',
    signalItems: [
      'Framework direction is public and documentation-first.',
      'Verified public entry point: GitHub repository and this docs site.',
      'Team/contact materials are published only when they are ready and real.',
    ],
    snapshotTitle: 'Current public snapshot',
    snapshotItems: [
      {label: 'Positioning', value: 'AI-driven full-stack development framework'},
      {label: 'System model', value: 'Host / Workspace / Plugins / OpenSpec'},
      {label: 'Reference stack', value: 'GoFrame + Vue 3 + Vben + Docusaurus'},
    ],
    principlesKicker: 'Project principles',
    principlesTitle: 'What this public site is trying to make explicit',
    principles: [
      {
        title: 'Readable architecture',
        body: 'The project is framed around explicit layers and clear contracts so contributors can inspect how the system is meant to evolve.',
      },
      {
        title: 'Truthful public surface',
        body: 'The site does not invent fake community channels, fake product screens, or placeholder contacts that look official but are not verified.',
      },
      {
        title: 'Spec-driven delivery',
        body: 'OpenSpec is treated as part of the engineering system, not as a side note, because AI output needs durable review and archive paths.',
      },
      {
        title: 'Composable extension model',
        body: 'Plugins exist to expand capability while keeping the host understandable, replaceable, and easier to govern in an open repo.',
      },
    ],
    mapKicker: 'Public map',
    mapTitle: 'How to navigate LinaPro as an open-source project',
    mapIntro:
      'If you are evaluating the project, these are the three surfaces that matter first: the architecture, the workflow, and the repository-facing extension model.',
    mapItems: [
      {
        name: 'Architecture docs',
        meta: 'System boundaries',
        body: 'Start with the platform layers, host responsibilities, workspace role, and extension seams.',
        href: '/docs/architecture',
        action: 'Read docs',
      },
      {
        name: 'OpenSpec workflow',
        meta: 'Change discipline',
        body: 'Understand how exploration, proposals, implementation, review, and archive fit into an AI-assisted delivery loop.',
        href: '/docs/openspec-workflow',
        action: 'View workflow',
      },
      {
        name: 'GitHub repository',
        meta: 'Canonical public entry',
        body: 'Use GitHub as the source of truth for code, issue tracking, release context, and future public collaboration signals.',
        href: 'https://github.com/linaproai/linapro',
        action: 'Open repo',
      },
    ],
    postureKicker: 'Release posture',
    postureTitle: 'What the public site commits to right now',
    postureIntro:
      'This site is intentionally conservative about what it claims. It should help readers understand the framework, not create synthetic trust.',
    postureItems: [
      'Show the real architecture before marketing broad promises.',
      'Publish only confirmed materials for contacts, teams, and ecosystem links.',
      'Keep documentation and public narrative aligned with repository reality.',
      'Make room for future accounts/community features without faking them in phase one.',
    ],
    footerTitle: 'Use the site as a map, then inspect the repo as the source of truth.',
    footerCopy:
      'LinaPro should feel like an engineering project first: documented, inspectable, and explicit about how AI fits into delivery.',
    footerPrimary: 'View GitHub',
    footerSecondary: 'Read docs',
  },
  'zh-CN': {
    eyebrow: '关于 LinaPro',
    title: '一套随着规模增长仍然保持可读性的开源 AI 工程系统。',
    body:
      'LinaPro 面向那些希望使用 AI 提升交付效率、但又不愿牺牲仓库结构、审查边界与系统清晰度的团队。这个公开站点的目标，就是用工程语言把这套模型讲清楚。',
    primary: '打开 GitHub',
    secondary: '阅读架构文档',
    signal: '公开运行说明',
    signalItems: [
      '项目方向已公开，并以文档为先。',
      '当前经过验证的公开入口：GitHub 仓库与本站文档。',
      '团队与联系方式只会在真实可确认后再发布。',
    ],
    snapshotTitle: '当前公开快照',
    snapshotItems: [
      {label: '项目定位', value: 'AI 驱动的全栈开发框架'},
      {label: '系统模型', value: '宿主 / 工作台 / 插件 / OpenSpec'},
      {label: '参考技术栈', value: 'GoFrame + Vue 3 + Vben + Docusaurus'},
    ],
    principlesKicker: '项目原则',
    principlesTitle: '这个公开站点希望明确表达的内容',
    principles: [
      {
        title: '可读的架构边界',
        body: '整个项目围绕显式分层与清晰契约组织，方便协作者理解系统应该如何演进，而不是靠隐式约定猜测。',
      },
      {
        title: '真实的公开入口面',
        body: '官网不会伪造社区渠道、伪造产品截图，也不会提供看似正式但未经验证的联系方式。',
      },
      {
        title: '规范驱动交付',
        body: 'OpenSpec 不是附属说明，而是工程系统的一部分，因为 AI 产出必须进入可审查、可归档的变更闭环。',
      },
      {
        title: '可组合的扩展模型',
        body: '插件的意义是增加能力，同时让宿主保持可理解、可替换、可治理，适合在开源仓库中长期演进。',
      },
    ],
    mapKicker: '公开导航',
    mapTitle: '如何从开源项目视角理解 LinaPro',
    mapIntro:
      '如果你正在评估这个项目，最先值得看的三个入口是：架构、工作流，以及面向仓库的扩展模型。',
    mapItems: [
      {
        name: '架构文档',
        meta: '系统边界',
        body: '先理解平台分层、宿主职责、工作台角色以及能力扩展的稳定边界。',
        href: '/docs/architecture',
        action: '查看文档',
      },
      {
        name: 'OpenSpec 工作流',
        meta: '变更纪律',
        body: '理解探索、提案、实现、审查、归档如何组成 AI 辅助研发中的完整交付闭环。',
        href: '/docs/openspec-workflow',
        action: '查看流程',
      },
      {
        name: 'GitHub 仓库',
        meta: '标准公开入口',
        body: '代码、问题跟踪、版本上下文以及未来协作信号，都应该以 GitHub 为准。',
        href: 'https://github.com/linaproai/linapro',
        action: '打开仓库',
      },
    ],
    postureKicker: '发布姿态',
    postureTitle: '当前公开站点真正承诺的内容',
    postureIntro:
      '这个站点对外表达是刻意克制的。它首先应该帮助读者理解框架，而不是制造虚假的信任感。',
    postureItems: [
      '先展示真实架构，再谈更大的对外承诺。',
      '团队、联系信息、生态链接只发布确认过的正式资料。',
      '保持文档、公开叙事与仓库事实一致。',
      '为未来账户和社区能力预留结构，但不在一期伪造它们。',
    ],
    footerTitle: '把站点当作地图，再把仓库当作最终事实来源。',
    footerCopy:
      'LinaPro 应该首先呈现为一个工程项目：有文档、可检查，并且明确说明 AI 在交付中扮演什么角色。',
    footerPrimary: '查看 GitHub',
    footerSecondary: '阅读文档',
  },
} as const;

export default function AboutPage(): ReactNode {
  const {i18n} = useDocusaurusContext();
  const locale = i18n.currentLocale as keyof typeof content;
  const copy = content[locale] ?? content.en;

  return (
    <Layout title={copy.eyebrow} description={copy.body}>
      <main className={styles.page}>
        <div className={styles.inner}>
          <section className={styles.hero}>
            <div className={styles.heroCopy}>
              <p className={styles.eyebrow}>{copy.eyebrow}</p>
              <h1>{copy.title}</h1>
              <p className={styles.heroText}>{copy.body}</p>
              <div className={styles.actions}>
                <Link
                  className={clsx('button', styles.primaryButton)}
                  href="https://github.com/linaproai/linapro">
                  {copy.primary}
                </Link>
                <Link className={clsx('button', styles.secondaryButton)} to="/docs/architecture">
                  {copy.secondary}
                </Link>
              </div>
            </div>

            <aside className={styles.signalPanel}>
              <div className={styles.panelHeader}>
                <span className={styles.panelLabel}>{copy.signal}</span>
                <span className={styles.panelBranch}>public/about</span>
              </div>
              <h2 className={styles.panelTitle}>
                {copy.snapshotTitle}
              </h2>
              <div className={styles.snapshotList}>
                {copy.snapshotItems.map((item) => (
                  <div className={styles.snapshotRow} key={item.label}>
                    <span>{item.label}</span>
                    <strong>{item.value}</strong>
                  </div>
                ))}
              </div>
              <div className={styles.signalList}>
                {copy.signalItems.map((item) => (
                  <div className={styles.signalLine} key={item}>
                    {item}
                  </div>
                ))}
              </div>
            </aside>
          </section>

          <section className={styles.section}>
            <div className={styles.sectionHeader}>
              <p className={styles.sectionKicker}>{copy.principlesKicker}</p>
              <h2>{copy.principlesTitle}</h2>
            </div>
            <div className={styles.principlesGrid}>
              {copy.principles.map((item) => (
                <article className={styles.principleCard} key={item.title}>
                  <h3>{item.title}</h3>
                  <p>{item.body}</p>
                </article>
              ))}
            </div>
          </section>

          <section className={clsx(styles.section, styles.mapSection)}>
            <div className={styles.sectionHeader}>
              <p className={styles.sectionKicker}>{copy.mapKicker}</p>
              <h2>{copy.mapTitle}</h2>
              <p>{copy.mapIntro}</p>
            </div>
            <div className={styles.mapGrid}>
              {copy.mapItems.map((item) => (
                <article className={styles.mapCard} key={item.name}>
                  <p className={styles.mapMeta}>{item.meta}</p>
                  <h3>{item.name}</h3>
                  <p>{item.body}</p>
                  <Link className={styles.inlineLink} href={item.href}>
                    {item.action}
                  </Link>
                </article>
              ))}
            </div>
          </section>

          <section className={clsx(styles.section, styles.postureSection)}>
            <article className={styles.postureCard}>
              <p className={styles.sectionKicker}>{copy.postureKicker}</p>
              <h2>{copy.postureTitle}</h2>
              <p className={styles.postureIntro}>{copy.postureIntro}</p>
              <ul className={styles.postureList}>
                {copy.postureItems.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            </article>
          </section>

          <section className={styles.closer}>
            <div>
              <h2>{copy.footerTitle}</h2>
              <p>{copy.footerCopy}</p>
            </div>
            <div className={styles.actions}>
              <Link
                className={clsx('button', styles.primaryButton)}
                href="https://github.com/linaproai/linapro">
                {copy.footerPrimary}
              </Link>
              <Link className={clsx('button', styles.secondaryButton)} to="/docs">
                {copy.footerSecondary}
              </Link>
            </div>
          </section>
        </div>
      </main>
    </Layout>
  );
}
