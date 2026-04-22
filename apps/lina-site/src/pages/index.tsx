import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

import styles from './index.module.css';

type SurfaceLink = {
  label: string;
  href: string;
  action: string;
  external?: boolean;
};

const content = {
  en: {
    heroEyebrow: 'Official website / docs workspace',
    heroTitle: 'A public engineering surface for AI-driven full-stack delivery.',
    heroDescription:
      'LinaPro presents its framework model the same way good engineering projects do: clear docs, explicit boundaries, public routes, and room to grow into a multi-app repository without hiding the architecture.',
    heroBadges: ['GoFrame runtime', 'OpenSpec workflow', 'Apps-first layout', 'Docs + blog + i18n'],
    primaryCta: 'Read the docs',
    secondaryCta: 'View GitHub',
    metrics: [
      {value: '04', label: 'explicit layers'},
      {value: '02', label: 'languages live'},
      {value: '01', label: 'official site app'},
      {value: '100%', label: 'public architecture narrative'},
    ],
    panelLabel: 'Monorepo snapshot',
    panelBranch: 'apps/lina-site',
    panelTitle: 'The official site now lives inside apps, ready for more product surfaces later.',
    treeItems: [
      {
        path: 'apps/lina-site',
        body: 'Official website, docs, blog, localized content, and brand-level public pages.',
      },
      {
        path: 'apps/<future-admin>',
        body: 'Reserved space for later management features, aligned with the repository shape used in LinaPro itself.',
      },
      {
        path: 'plugins/*',
        body: 'Capability modules stay separate so the host and apps do not collapse into one unreadable surface.',
      },
      {
        path: 'openspec/*',
        body: 'Specification, review, and archive flow stays visible instead of disappearing into chat history.',
      },
    ],
    commandTitle: 'Site commands',
    commands: ['make dev', 'pnpm typecheck', 'pnpm build'],
    featuresEyebrow: 'Framework highlights',
    featuresTitle: 'Clear public positioning before broader product expansion.',
    featuresIntro:
      'The homepage now follows a cleaner framework-site rhythm: headline first, strengths second, repository layout third, then deeper architecture sections for readers who want to evaluate the system seriously.',
    features: [
      {
        title: 'Docs-first public surface',
        body: 'The site behaves like a real engineering entry point instead of a disconnected marketing shell.',
      },
      {
        title: 'Apps-first repository shape',
        body: 'Putting the site in apps/lina-site makes future admin or operator products easy to add beside it.',
      },
      {
        title: 'Explicit extension seams',
        body: 'Host, workspace, plugins, and OpenSpec are presented as readable boundaries, not blurred framework magic.',
      },
      {
        title: 'Bilingual from the start',
        body: 'English and Simplified Chinese stay first-class so the official narrative works for both audiences.',
      },
      {
        title: 'Trust through structure',
        body: 'Public routes, stable docs, and visible repo organization create more trust than inflated product claims.',
      },
      {
        title: 'Ready for management apps',
        body: 'The site layout already leaves room for later operational and management interfaces under the same apps directory.',
      },
    ],
    repoEyebrow: 'Repository layout',
    repoTitle: 'One official site app now, a broader app family later.',
    repoIntro:
      'This mirrors the structure you referenced from the LinaPro repository: keep the official site as one app, then grow other management-oriented surfaces beside it without rewriting the public website later.',
    repoLeadTitle: 'Why this matters',
    repoLeadBody:
      'The website stops being a special case at the repository root. It becomes one application inside a clearer workspace model, which makes future expansion more natural and easier to maintain.',
    repoCards: [
      {
        path: 'apps/lina-site',
        title: 'Official brand and docs entry',
        body: 'Landing page, docs, blog, about page, and future public account routes remain together in one app boundary.',
      },
      {
        path: 'apps/*',
        title: 'Room for future management features',
        body: 'Admin consoles, control panels, or operator tools can land beside the official site instead of forcing a repo reshuffle later.',
      },
      {
        path: 'static + i18n',
        title: 'Localized assets stay close to the site',
        body: 'Brand graphics, translated docs, and public content remain co-located with the website application that actually serves them.',
      },
      {
        path: 'root workspace',
        title: 'Single pnpm entrypoint',
        body: 'Root scripts now delegate into the site app, which is a better foundation for a multi-app repository than a single flat package.',
      },
    ],
    architectureEyebrow: 'Operating model',
    architectureTitle: 'Four layers, one readable delivery model.',
    architectureIntro:
      'The official site should explain the framework with the same clarity the framework expects from the codebase itself.',
    layers: [
      {
        index: '01',
        name: 'Core host service',
        meta: 'GoFrame runtime foundation',
        body: 'Shared contracts, auth, lifecycle, governance, and plugin boundaries belong here.',
      },
      {
        index: '02',
        name: 'Management workspace',
        meta: 'Operational product surface',
        body: 'The operator-facing UI grows as a dedicated app instead of being confused with the official site.',
      },
      {
        index: '03',
        name: 'Plugin runtime',
        meta: 'Official + custom modules',
        body: 'Extension capability stays modular so teams can swap or audit features without rewriting the core.',
      },
      {
        index: '04',
        name: 'OpenSpec workflow',
        meta: 'Explore -> Propose -> Implement -> Review',
        body: 'Specs, review gates, and archive paths keep AI-assisted delivery tied to repository truth.',
      },
    ],
    surfacesEyebrow: 'Day-one surfaces',
    surfacesTitle: 'The official site already covers the public essentials.',
    surfacesIntro:
      'Before the management apps arrive, the website still needs to do several jobs well: explain the project, expose docs, host updates, and keep future routes stable.',
    surfaces: [
      {
        label: 'Docs',
        href: '/docs/architecture',
        action: 'Open architecture',
        body: 'Start with the system model, layers, and public implementation boundaries.',
      },
      {
        label: 'Blog',
        href: '/blog',
        action: 'Read writing',
        body: 'Use the blog for release notes, architecture notes, and framework evolution stories.',
      },
      {
        label: 'About',
        href: '/about',
        action: 'See project posture',
        body: 'Explain what is public, what is intentionally reserved, and what the project truly commits to right now.',
      },
      {
        label: 'GitHub',
        href: 'https://github.com/linaproai/linapro',
        action: 'Open repository',
        external: true,
        body: 'The repository remains the source of truth for code, issues, releases, and future collaboration signals.',
      },
    ] satisfies Array<SurfaceLink & {body: string}>,
    footerEyebrow: 'Next step',
    footerTitle: 'Read the docs, inspect the repo, then grow more apps beside the official site.',
    footerBody:
      'The new structure gives the website a cleaner home today and a better expansion path for future management functionality tomorrow.',
    footerPrimary: 'Explore architecture',
    footerSecondary: 'Read OpenSpec workflow',
  },
  'zh-CN': {
    heroEyebrow: '官方网站 / 文档工作区',
    heroTitle: '把 AI 驱动的全栈工程体系，先讲清楚，再逐步扩展成多应用仓库。',
    heroDescription:
      'LinaPro 官网现在按照更接近成熟框架站点的方式组织：先有清晰文档、显式边界、稳定公开路由，再自然扩展为包含更多应用的仓库，而不是把官网长期放在根目录里单独维护。',
    heroBadges: ['GoFrame 运行时', 'OpenSpec 工作流', 'Apps-first 结构', '文档 + 博客 + 双语'],
    primaryCta: '阅读文档',
    secondaryCta: '查看 GitHub',
    metrics: [
      {value: '04', label: '显式层次'},
      {value: '02', label: '已上线语言'},
      {value: '01', label: '官网站点应用'},
      {value: '100%', label: '公开架构叙事'},
    ],
    panelLabel: 'Monorepo 快照',
    panelBranch: 'apps/lina-site',
    panelTitle: '官网站点已经进入 apps 目录，后续可以在旁边继续扩展更多产品入口面。',
    treeItems: [
      {
        path: 'apps/lina-site',
        body: '官方网站、文档、博客、多语言内容，以及品牌层面的公开页面都集中在这里。',
      },
      {
        path: 'apps/<future-admin>',
        body: '为后续管理功能预留稳定位置，整体结构更接近 LinaPro 主仓库的组织方式。',
      },
      {
        path: 'plugins/*',
        body: '能力模块继续保持独立，让宿主与应用层不会混成一个不可读的大表面。',
      },
      {
        path: 'openspec/*',
        body: '规范、审查、归档链路保持可见，而不是消失在会话历史里。',
      },
    ],
    commandTitle: '站点命令',
    commands: ['make dev', 'pnpm typecheck', 'pnpm build'],
    featuresEyebrow: '框架亮点',
    featuresTitle: '先把公开官网做好，再承接更大的产品扩展。',
    featuresIntro:
      '新的首页更接近 GoFrame 这类框架官网的阅读节奏：先看主标题，再看能力亮点，再看仓库结构和系统分层，方便首次访问者快速理解项目定位。',
    features: [
      {
        title: '文档优先的公开入口面',
        body: '官网首先是一个真正可用的工程入口，而不是和文档脱节的营销壳子。',
      },
      {
        title: 'Apps-first 的仓库组织',
        body: '把站点放进 apps/lina-site，后续增加后台或管理产品时不需要重新整理仓库。',
      },
      {
        title: '显式的扩展边界',
        body: '宿主、工作台、插件、OpenSpec 都以可读边界呈现，而不是隐藏在框架魔法里。',
      },
      {
        title: '从一开始就支持双语',
        body: '英文与简体中文都是一等公民，让官网叙事面向更真实的使用群体。',
      },
      {
        title: '通过结构建立信任',
        body: '稳定路由、清晰文档、明确仓库布局，比夸张宣传更能建立可信度。',
      },
      {
        title: '为管理功能预留空间',
        body: '官网结构已经为后续运维或管理型应用做好位置准备，可以自然继续扩展。',
      },
    ],
    repoEyebrow: '仓库布局',
    repoTitle: '先有一个官网应用，再逐步长成一组应用。',
    repoIntro:
      '这和你提到的 LinaPro 仓库组织方式是一致的：先把官网作为一个 app 放进 apps 目录，后续再把管理功能或操作台类入口面并列放进去，而不是将来再整体搬迁。',
    repoLeadTitle: '这样做的价值',
    repoLeadBody:
      '官网不再是根目录下的特例，而是工作区中的一个正式应用。这样后续扩展更自然，维护边界也更清晰。',
    repoCards: [
      {
        path: 'apps/lina-site',
        title: '品牌官网与文档主入口',
        body: '首页、文档、博客、About 页面，以及未来的公开账户路由都集中在一个应用边界里。',
      },
      {
        path: 'apps/*',
        title: '为后续管理功能留位',
        body: '后台、控制台、运营面板等都可以直接并列增加，无需再打断官网工程结构。',
      },
      {
        path: 'static + i18n',
        title: '站点素材与多语言就近管理',
        body: '品牌图形、翻译文档与公开内容，都与真正消费它们的站点应用放在一起。',
      },
      {
        path: 'root workspace',
        title: '统一的 pnpm 工作区入口',
        body: '根目录脚本现在转发到站点应用，为未来多应用仓库提供更自然的基础。',
      },
    ],
    architectureEyebrow: '运行模型',
    architectureTitle: '四层结构，一套可读的交付模型。',
    architectureIntro:
      '官网应该用和框架本身一致的方式来解释框架：边界清楚、层次分明、可公开阅读。',
    layers: [
      {
        index: '01',
        name: '核心宿主服务',
        meta: 'GoFrame 运行时底座',
        body: '共享契约、认证权限、生命周期、治理能力和插件边界都在这里。',
      },
      {
        index: '02',
        name: '管理工作台',
        meta: '面向运维与操作的产品入口',
        body: '运营或管理型 UI 会以独立 app 的形式增长，而不是和官网混在一起。',
      },
      {
        index: '03',
        name: '插件运行时',
        meta: '官方能力 + 自定义模块',
        body: '扩展能力保持模块化，方便替换、审计和按需启停。',
      },
      {
        index: '04',
        name: 'OpenSpec 工作流',
        meta: 'Explore -> Propose -> Implement -> Review',
        body: '让 AI 参与交付时，规范、审查与归档链路始终对仓库保持可见。',
      },
    ],
    surfacesEyebrow: '一期公开入口',
    surfacesTitle: '即使管理应用还没加入，官网也已经覆盖了核心公开职责。',
    surfacesIntro:
      '在后续管理功能上线之前，官方网站至少要把这些事情做好：解释项目、承载文档、发布更新，并为未来路由保持稳定入口。',
    surfaces: [
      {
        label: 'Docs',
        href: '/docs/architecture',
        action: '查看架构',
        body: '从系统模型、层次划分和公开实现边界开始理解 LinaPro。',
      },
      {
        label: 'Blog',
        href: '/blog',
        action: '阅读博客',
        body: '适合发布版本说明、架构文章和框架演进记录。',
      },
      {
        label: 'About',
        href: '/about',
        action: '查看项目姿态',
        body: '说明哪些内容已经公开、哪些能力仍是预留，以及项目当前真正承诺的边界。',
      },
      {
        label: 'GitHub',
        href: 'https://github.com/linaproai/linapro',
        action: '打开仓库',
        external: true,
        body: '代码、Issue、版本和后续协作信号，仍然都应该以仓库为准。',
      },
    ] satisfies Array<SurfaceLink & {body: string}>,
    footerEyebrow: '下一步',
    footerTitle: '先从文档和仓库理解系统，再把更多应用自然加到官网旁边。',
    footerBody:
      '新的结构让官网今天就有更清晰的位置，也让后续管理功能明天可以更顺畅地进入 apps 目录。',
    footerPrimary: '查看架构',
    footerSecondary: '阅读 OpenSpec 工作流',
  },
} as const;

export default function Home(): ReactNode {
  const wordmarkUrl = useBaseUrl('/img/linapro-wordmark.svg');
  const {i18n} = useDocusaurusContext();
  const locale = i18n.currentLocale as keyof typeof content;
  const copy = content[locale] ?? content.en;

  return (
    <Layout title={copy.heroEyebrow} description={copy.heroDescription}>
      <main className={styles.page}>
        <div className={styles.container}>
          <section className={styles.hero}>
            <div className={styles.heroCopy}>
              <p className={styles.heroEyebrow}>{copy.heroEyebrow}</p>
              <img className={styles.wordmark} src={wordmarkUrl} alt="LinaPro" />
              <h1 className={styles.heroTitle}>
                {copy.heroTitle}
              </h1>
              <p className={styles.heroDescription}>{copy.heroDescription}</p>
              <div className={styles.badgeRow}>
                {copy.heroBadges.map((badge) => (
                  <span className={styles.badge} key={badge}>
                    {badge}
                  </span>
                ))}
              </div>
              <div className={styles.actionRow}>
                <Link className={clsx('button', styles.primaryButton)} to="/docs">
                  {copy.primaryCta}
                </Link>
                <Link
                  className={clsx('button', styles.secondaryButton)}
                  href="https://github.com/linaproai/linapro">
                  {copy.secondaryCta}
                </Link>
              </div>
              <div className={styles.metricGrid}>
                {copy.metrics.map((item) => (
                  <article className={styles.metricCard} key={item.label}>
                    <strong>{item.value}</strong>
                    <span>{item.label}</span>
                  </article>
                ))}
              </div>
            </div>

            <aside className={styles.heroPanel}>
              <div className={styles.panelHeader}>
                <span className={styles.panelLabel}>{copy.panelLabel}</span>
                <span className={styles.panelBranch}>{copy.panelBranch}</span>
              </div>
              <h2 className={styles.panelTitle}>
                {copy.panelTitle}
              </h2>
              <div className={styles.treeList}>
                {copy.treeItems.map((item) => (
                  <div className={styles.treeItem} key={item.path}>
                    <code className={styles.pathChip}>{item.path}</code>
                    <p>{item.body}</p>
                  </div>
                ))}
              </div>
              <div className={styles.commandBlock}>
                <p className={styles.commandTitle}>{copy.commandTitle}</p>
                <div className={styles.commandList}>
                  {copy.commands.map((command) => (
                    <div className={styles.commandLine} key={command}>
                      <span>$</span>
                      <code>{command}</code>
                    </div>
                  ))}
                </div>
              </div>
            </aside>
          </section>

          <section className={styles.section}>
            <div className={styles.sectionHeader}>
              <p className={styles.sectionEyebrow}>{copy.featuresEyebrow}</p>
              <h2 className={styles.sectionTitle}>
                {copy.featuresTitle}
              </h2>
              <p className={styles.sectionIntro}>{copy.featuresIntro}</p>
            </div>
            <div className={styles.featureGrid}>
              {copy.features.map((item) => (
                <article className={styles.featureCard} key={item.title}>
                  <h3>{item.title}</h3>
                  <p>{item.body}</p>
                </article>
              ))}
            </div>
          </section>

          <section className={styles.section}>
            <div className={styles.repoLayout}>
              <article className={styles.repoLead}>
                <p className={styles.sectionEyebrow}>{copy.repoEyebrow}</p>
                <h2 className={styles.sectionTitle}>
                  {copy.repoTitle}
                </h2>
                <p className={styles.sectionIntro}>{copy.repoIntro}</p>
                <div className={styles.leadCallout}>
                  <strong>{copy.repoLeadTitle}</strong>
                  <p>{copy.repoLeadBody}</p>
                </div>
              </article>
              <div className={styles.repoGrid}>
                {copy.repoCards.map((item) => (
                  <article className={styles.repoCard} key={item.title}>
                    <code className={styles.pathChip}>{item.path}</code>
                    <h3>{item.title}</h3>
                    <p>{item.body}</p>
                  </article>
                ))}
              </div>
            </div>
          </section>

          <section className={styles.section}>
            <div className={styles.sectionHeader}>
              <p className={styles.sectionEyebrow}>{copy.architectureEyebrow}</p>
              <h2 className={styles.sectionTitle}>
                {copy.architectureTitle}
              </h2>
              <p className={styles.sectionIntro}>{copy.architectureIntro}</p>
            </div>
            <div className={styles.layerGrid}>
              {copy.layers.map((item) => (
                <article className={styles.layerCard} key={item.name}>
                  <span className={styles.layerIndex}>{item.index}</span>
                  <h3>{item.name}</h3>
                  <p className={styles.layerMeta}>{item.meta}</p>
                  <p>{item.body}</p>
                </article>
              ))}
            </div>
          </section>

          <section className={styles.section}>
            <div className={styles.sectionHeader}>
              <p className={styles.sectionEyebrow}>{copy.surfacesEyebrow}</p>
              <h2 className={styles.sectionTitle}>
                {copy.surfacesTitle}
              </h2>
              <p className={styles.sectionIntro}>{copy.surfacesIntro}</p>
            </div>
            <div className={styles.surfaceGrid}>
              {copy.surfaces.map((item) => (
                <article className={styles.surfaceCard} key={item.label}>
                  <span className={styles.surfaceLabel}>{item.label}</span>
                  <p>{item.body}</p>
                  <Link className={styles.inlineLink} {...(item.external ? {href: item.href} : {to: item.href})}>
                    {item.action}
                  </Link>
                </article>
              ))}
            </div>
          </section>

          <section className={styles.ctaSection}>
            <p className={styles.sectionEyebrow}>{copy.footerEyebrow}</p>
            <h2 className={styles.sectionTitle}>
              {copy.footerTitle}
            </h2>
            <p className={styles.sectionIntro}>{copy.footerBody}</p>
            <div className={styles.actionRow}>
              <Link className={clsx('button', styles.primaryButton)} to="/docs/architecture">
                {copy.footerPrimary}
              </Link>
              <Link className={clsx('button', styles.secondaryButton)} to="/docs/openspec-workflow">
                {copy.footerSecondary}
              </Link>
            </div>
          </section>
        </div>
      </main>
    </Layout>
  );
}
