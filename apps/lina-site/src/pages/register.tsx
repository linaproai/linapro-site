import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';

import styles from './pageShell.module.css';

const content = {
  en: {
    eyebrow: 'Reserved entry',
    title: 'Registration is planned, but not shipped in phase one.',
    body: 'The public site reserves this route for future account onboarding, but the first release focuses on the official site, docs, blog, and public framework narrative.',
    status: 'placeholder',
    panelTitle: 'What comes first',
    lines: [
      'Official website structure',
      'Public documentation and blog',
      'Future-ready route reserved for user system integration',
    ],
    primary: 'Explore architecture',
    secondary: 'Read About',
  },
  'zh-CN': {
    eyebrow: '预留入口',
    title: '注册能力已预留，但不属于一期交付。',
    body: '公开站点先保留这个路由，便于后续对接账户体系；第一阶段的重点仍然是官网、文档、博客和对外叙事本身。',
    status: '占位中',
    panelTitle: '当前优先级',
    lines: [
      '官方网站结构',
      '公开文档与博客',
      '为未来用户系统集成预留稳定路由',
    ],
    primary: '查看架构',
    secondary: '阅读 About',
  },
} as const;

export default function RegisterPage(): ReactNode {
  const {i18n} = useDocusaurusContext();
  const locale = i18n.currentLocale as keyof typeof content;
  const copy = content[locale] ?? content.en;

  return (
    <Layout title={copy.eyebrow} description={copy.body}>
      <main className={styles.shell}>
        <div className={styles.inner}>
          <section className={styles.hero}>
            <div>
              <p className={styles.eyebrow}>{copy.eyebrow}</p>
              <h1>{copy.title}</h1>
              <p>{copy.body}</p>
              <div className={styles.actions}>
                <Link className={clsx('button', styles.primaryButton)} to="/docs/architecture">
                  {copy.primary}
                </Link>
                <Link className={clsx('button', styles.secondaryButton)} to="/about">
                  {copy.secondary}
                </Link>
              </div>
            </div>
            <aside className={styles.panel}>
              <span className={styles.status}>{copy.status}</span>
              <h2>{copy.panelTitle}</h2>
              {copy.lines.map((line) => (
                <div className={styles.infoLine} key={line}>
                  {line}
                </div>
              ))}
            </aside>
          </section>
        </div>
      </main>
    </Layout>
  );
}
