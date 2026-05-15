import Link from '@docusaurus/Link';
import Translate, {translate} from '@docusaurus/Translate';
import useBaseUrl from '@docusaurus/useBaseUrl';
import CodeBlock from '@theme/CodeBlock';
import Layout from '@theme/Layout';
import {useEffect, useState} from 'react';

const REPO_URL = 'https://github.com/linaproai/linapro';
const DOCS_URL = '/quick/overview';

function HomepageHeader() {
    const githubImage = useBaseUrl('/img/github.svg');
    const logoImage = useBaseUrl('/img/linapro-logo.png');
    return (
        <header className="home-section home-section--hero">
            <div className="container logo-container">
                <div>
                    <img
                        src={logoImage}
                        className="logo"
                        alt={translate({
                            id: 'home.hero.logoAlt',
                            message: 'LinaPro',
                            description: 'Alt text for the LinaPro logo in the hero',
                        })}
                    />
                </div>

                <h1 className="hero-title">
                    <Translate id="home.hero.title" description="Hero h1">
                        AI-Native Full-Stack Framework
                    </Translate>
                </h1>

                <p className="hero-lead">
                    <Translate id="home.hero.lead" description="Hero lead paragraph">
                        Building production-grade full-stack applications with AI, while keeping architecture, testing, and governance under control.
                    </Translate>
                </p>

                <div className="hero-cta">
                    <Link
                        className="button button--primary button--md"
                        to={DOCS_URL}
                        style={{width: '200px'}}
                    >
                        <Translate id="home.hero.cta.start" description="Hero primary CTA">
                            Get Started →
                        </Translate>
                    </Link>
                    <Link
                        className="button button--secondary button--md hover:bg-gray-200"
                        to={REPO_URL}
                        style={{
                            width: '200px',
                            paddingLeft: '50px',
                            backgroundImage: `url(${githubImage})`,
                            backgroundRepeat: 'no-repeat',
                            backgroundPosition: '45px center',
                        }}
                    >
                        <Translate id="home.hero.cta.github" description="Hero secondary CTA">
                            GitHub →
                        </Translate>
                    </Link>
                </div>
            </div>
        </header>
    );
}

const layers = [
    {
        id: 'core',
        moduleKey: 'apps/lina-core',
        title: (
            <Translate id="home.layer.core.title" description="Layer 1 title">
                Core Host Service
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.core.desc" description="Layer 1 description">
                A general-purpose Go backend runtime that handles API contracts, service governance, auth, permissions, and the plugin lifecycle out of the box.
            </Translate>
        ),
    },
    {
        id: 'vben',
        moduleKey: 'apps/lina-vben',
        title: (
            <Translate id="home.layer.vben.title" description="Layer 2 title">
                Management Workspace
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.vben.desc" description="Layer 2 description">
                A production-ready Vue 3 workspace that ships as the reference UI for every built-in capability — extend it, or swap it out entirely.
            </Translate>
        ),
    },
    {
        id: 'plugins',
        moduleKey: 'apps/lina-plugins',
        title: (
            <Translate id="home.layer.plugins.title" description="Layer 3 title">
                Dual-mode Plugin System
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.plugins.desc" description="Layer 3 description">
                Hot-loadable source plugins and WASM runtime plugins, each sandboxed with namespaced database and filesystem access — extend or replace anything without touching the host.
            </Translate>
        ),
    },
    {
        id: 'openspec',
        moduleKey: 'openspec/',
        title: (
            <Translate id="home.layer.openspec.title" description="Layer 4 title">
                AI-native R&D Workflow
            </Translate>
        ),
        desc: (
            <Translate id="home.layer.openspec.desc" description="Layer 4 description">
                A spec-first workflow that keeps AI, humans, and the codebase in lockstep across every iteration.
            </Translate>
        ),
    },
];

function Layers() {
    return (
        <section className="home-section home-section--layers">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.layers.title" description="Layers section title">
                        Composable Full-stack Architecture
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.layers.lead" description="Layers section lead">
                        A composable full-stack architecture where each layer can evolve independently.
                    </Translate>
                </p>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-5 mt-8">
                    {layers.map((layer, idx) => (
                        <div key={layer.id} className="card layer-card p-8 box-border">
                            <div className="layer-number">{`0${idx + 1}`}</div>
                            <h3 className="layer-title">{layer.title}</h3>
                            <code className="layer-key">{layer.moduleKey}</code>
                            <p className="layer-desc">{layer.desc}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

const workflowSteps = [
    {
        id: 'explore',
        title: (
            <Translate id="home.workflow.explore.title" description="Workflow stage 1">
                Explore
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.explore.desc" description="Workflow stage 1 desc">
                Map out the problem space and weigh the options before committing to one.
            </Translate>
        ),
    },
    {
        id: 'propose',
        title: (
            <Translate id="home.workflow.propose.title" description="Workflow stage 2">
                Propose
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.propose.desc" description="Workflow stage 2 desc">
                Turn intent into an incremental spec, anchored to clear acceptance criteria.
            </Translate>
        ),
    },
    {
        id: 'implement',
        title: (
            <Translate id="home.workflow.implement.title" description="Workflow stage 3">
                Implement
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.implement.desc" description="Workflow stage 3 desc">
                AI builds against the spec and writes the matching E2E tests alongside it.
            </Translate>
        ),
    },
    {
        id: 'review',
        title: (
            <Translate id="home.workflow.review.title" description="Workflow stage 4">
                Review
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.review.desc" description="Workflow stage 4 desc">
                Humans sign off on the spec, the code, and the tests at the gates that matter.
            </Translate>
        ),
    },
    {
        id: 'archive',
        title: (
            <Translate id="home.workflow.archive.title" description="Workflow stage 5">
                Archive
            </Translate>
        ),
        desc: (
            <Translate id="home.workflow.archive.desc" description="Workflow stage 5 desc">
                The verified spec joins the baseline, so the next iteration starts on solid ground.
            </Translate>
        ),
    },
];

function Workflow() {
    return (
        <section className="home-section home-section--workflow">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.workflow.title" description="Workflow section title">
                        Spec-driven AI development loop
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.workflow.lead" description="Workflow section lead">
                        A spec-driven development loop that keeps AI output aligned with requirements, code, and tests.
                    </Translate>
                </p>
                <div className="workflow-steps">
                    {workflowSteps.map((step, i) => (
                        <div key={step.id} className="workflow-step">
                            <div className="workflow-step-number">{`STAGE 0${i + 1}`}</div>
                            <h4>{step.title}</h4>
                            <p>{step.desc}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

const IconSparkles = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 3l1.6 4.4L18 9l-4.4 1.6L12 15l-1.6-4.4L6 9l4.4-1.6z" />
        <path d="M19 15l.7 1.8L21.5 17.5l-1.8.7L19 20l-.7-1.8-1.8-.7 1.8-.7z" />
        <path d="M5 15l.5 1.3L6.8 16.8l-1.3.5L5 18.5l-.5-1.3L3.3 16.8l1.3-.5z" />
    </svg>
);

const IconContract = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <path d="M6 3h8l4 4v14H6z" />
        <path d="M14 3v5h5" />
        <path d="M9 12h6" />
        <path d="M9 16h3" />
        <path d="M4 8H2v11a2 2 0 0 0 2 2h2" />
    </svg>
);

const IconBlocks = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="7" rx="1" />
        <rect x="14" y="3" width="7" height="7" rx="1" />
        <rect x="3" y="14" width="7" height="7" rx="1" />
        <rect x="14" y="14" width="7" height="7" rx="1" />
    </svg>
);

const IconPlug = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 2v6" />
        <path d="M15 2v6" />
        <path d="M5 8h14v3a5 5 0 0 1-5 5h-4a5 5 0 0 1-5-5z" />
        <path d="M12 16v6" />
    </svg>
);

const IconGlobe = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="9" />
        <path d="M3 12h18" />
        <path d="M12 3c2.3 2.4 3.5 5.4 3.5 9S14.3 18.6 12 21" />
        <path d="M12 3c-2.3 2.4-3.5 5.4-3.5 9S9.7 18.6 12 21" />
    </svg>
);

const IconNetwork = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="5" r="2.2" />
        <circle cx="5" cy="19" r="2.2" />
        <circle cx="19" cy="19" r="2.2" />
        <path d="M12 7v3M12 10l-5.5 7M12 10l5.5 7" />
    </svg>
);

const IconSkillLib = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <rect x="7" y="7" width="10" height="10" rx="1" />
        <path d="M10 7V5M14 7V5M10 19v-2M14 19v-2M7 10H5M7 14H5M19 10h-2M19 14h-2" />
        <circle cx="12" cy="12" r="2" />
    </svg>
);

const IconCheckCircle = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="9" />
        <path d="M8 12l3 3 5-6" />
    </svg>
);

const IconTenants = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="8" cy="7" r="2.5" />
        <path d="M3 19c0-3 2-5 5-5s5 2 5 5" />
        <circle cx="17" cy="7" r="2.5" />
        <path d="M14 19c0-2.5 1.3-4.3 3-4.8" />
        <path d="M20 14.5c1 .8 1.7 2.1 1.8 3.5" />
        <path d="M14 19h7" />
    </svg>
);

const strengths = [
    {
        id: 'workflow',
        Icon: IconSparkles,
        title: (
            <Translate id="home.strength.workflow.title" description="Strength 1 title">
                AI-native R&D Workflow
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.workflow.desc" description="Strength 1 desc">
                AI advances from exploration through archiving against verified specs, keeping requirements, code, and tests continuously aligned.
            </Translate>
        ),
    },
    {
        id: 'skills',
        Icon: IconSkillLib,
        title: (
            <Translate id="home.strength.skills.title" description="Strength skills title">
                Rich AI Skills Ecosystem
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.skills.desc" description="Strength skills desc">
                A built-in library of AI skills for code assistance, review, test generation, and workflow automation — composable and extensible to any project.
            </Translate>
        ),
    },
    {
        id: 'coupling',
        Icon: IconBlocks,
        title: (
            <Translate id="home.strength.coupling.title" description="Strength 2 title">
                Module-level Loose Coupling
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.coupling.desc" description="Strength 2 desc">
                Modules depend on interfaces instead of hard dependencies, so each can be enabled, replaced, or retired without full-stack rewrites.
            </Translate>
        ),
    },
    {
        id: 'plugins',
        Icon: IconPlug,
        title: (
            <Translate id="home.strength.plugins.title" description="Strength 3 title">
                Dual-mode Pluggable Runtime
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.plugins.desc" description="Strength 3 desc">
                Source plugins handle deep extensions; WASM plugins hot-load at runtime — both run in isolated sandboxes with namespaced resource access.
            </Translate>
        ),
    },
    {
        id: 'contract',
        Icon: IconContract,
        title: (
            <Translate id="home.strength.contract.title" description="Strength contract title">
                Unified Full-stack Contracts
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.contract.desc" description="Strength contract desc">
                Service APIs, OpenAPI docs, and permission metadata share one model with the workspace, so all interfaces stay documented and testable.
            </Translate>
        ),
    },
    {
        id: 'i18n',
        Icon: IconGlobe,
        title: (
            <Translate id="home.strength.i18n.title" description="Strength localization title">
                Framework-level I18n Governance
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.i18n.desc" description="Strength localization desc">
                The host, plugins, and API docs share translation resources, with file-based language packs, missing-key checks, and runtime caching.
            </Translate>
        ),
    },
    {
        id: 'distributed',
        Icon: IconNetwork,
        title: (
            <Translate id="home.strength.distributed.title" description="Strength 5 title">
                Native Distributed Architecture
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.distributed.desc" description="Strength 5 desc">
                Permission sync, locks, and caching are cluster-aware by default, letting a single node scale to a full cluster without code changes.
            </Translate>
        ),
    },
    {
        id: 'multitenant',
        Icon: IconTenants,
        title: (
            <Translate id="home.strength.multitenant.title" description="Strength multitenant title">
                Native Multi-tenant Support
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.multitenant.desc" description="Strength multitenant desc">
                Tenant middleware, identity context, and governance hooks are built into the host core, with automatic fallback to single-tenant mode.
            </Translate>
        ),
    },
    {
        id: 'builtin',
        Icon: IconCheckCircle,
        title: (
            <Translate id="home.strength.builtin.title" description="Strength 6 title">
                Production-ready Out of the Box
            </Translate>
        ),
        desc: (
            <Translate id="home.strength.builtin.desc" description="Strength 6 desc">
                JWT, declarative RBAC, forced logout, and IP/device auditing are built in, with permission changes applied in seconds from day one.
            </Translate>
        ),
    },
];

function Strengths() {
    return (
        <section className="home-section home-section--features">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.strengths.title" description="Strengths section title">
                        Framework Highlights
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.strengths.lead" description="Strengths section lead">
                        Compound AI productivity with every iteration, while keeping delivery quality verifiable and adaptable at scale.
                    </Translate>
                </p>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-5 mt-8">
                    {strengths.map(({id, Icon, title, desc}) => (
                        <div key={id} className="card p-8 box-border">
                            <div className="feature-icon">
                                <Icon />
                            </div>
                            <h3 className="strength-title">{title}</h3>
                            <p className="strength-desc">{desc}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

const previewSlots = [
    {
        id: 'i18n',
        img: '/img/preview/linapro-i18n.webp',
        caption: (
            <Translate id="home.modules.preview.i18n" description="Screenshot slot: i18n governance">
                Framework-level I18n Governance
            </Translate>
        ),
    },
    {
        id: 'plugin',
        img: '/img/preview/linapro-plugin.webp',
        caption: (
            <Translate id="home.modules.preview.plugin" description="Screenshot slot: plugin management">
                Plugin Management
            </Translate>
        ),
    },
    {
        id: 'apidoc',
        img: '/img/preview/linapro-apidoc.webp',
        caption: (
            <Translate id="home.modules.preview.apidoc" description="Screenshot slot: API docs">
                Built-in API Docs
            </Translate>
        ),
    },
    {
        id: 'menu',
        img: '/img/preview/linapro-menu.webp',
        caption: (
            <Translate id="home.modules.preview.menu" description="Screenshot slot: menu management">
                Menu Management
            </Translate>
        ),
    },
];

function Modules() {
    const imgUrls = [
        useBaseUrl(previewSlots[0].img),
        useBaseUrl(previewSlots[1].img),
        useBaseUrl(previewSlots[2].img),
        useBaseUrl(previewSlots[3].img),
    ];
    const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

    useEffect(() => {
        if (!lightboxSrc) return;
        const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setLightboxSrc(null); };
        document.addEventListener('keydown', onKey);
        return () => document.removeEventListener('keydown', onKey);
    }, [lightboxSrc]);

    return (
        <>
            {lightboxSrc && (
                <div className="preview-lightbox" onClick={() => setLightboxSrc(null)}>
                    <img src={lightboxSrc} className="preview-lightbox-img" />
                </div>
            )}
            <section className="home-section home-section--modules">
                <div className="container">
                    <h2 className="section-title">
                        <Translate id="home.modules.title" description="Modules section title">
                            A workspace ready to use on day one
                        </Translate>
                    </h2>
                    <p className="section-lead">
                        <Translate id="home.modules.lead" description="Modules section lead">
                            Start with a ready-to-use management workspace, then extend or replace modules as your product grows.
                        </Translate>
                    </p>
                    <div className="modules-preview-grid">
                        {previewSlots.map((slot, idx) => (
                            <div key={slot.id} className="modules-preview-card">
                                <img
                                    src={imgUrls[idx]}
                                    alt={slot.id}
                                    className="modules-preview-img"
                                    onClick={() => setLightboxSrc(imgUrls[idx])}
                                />
                            </div>
                        ))}
                    </div>
                </div>
            </section>
        </>
    );
}

const quickStartBlocks = [
    {
        id: 'install',
        label: 'INSTALL',
        body: 'git clone --depth 1 https://github.com/linaproai/linapro.git linapro',
    },
    {
        id: 'init',
        label: 'INITIALIZE',
        body: 'make init confirm=init\nmake mock confirm=mock   # optional demo data',
    },
    {
        id: 'dev',
        label: 'DEVELOP',
        body: 'make dev\n# Workspace: http://localhost:5666\n# API:       http://localhost:8080',
    },
];

function QuickStart() {
    return (
        <section className="home-section home-section--quickstart">
            <div className="container">
                <h2 className="section-title">
                    <Translate id="home.quickstart.title" description="QuickStart section title">
                        Up and running in three commands
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.quickstart.lead" description="QuickStart section lead">
                        Install, initialize, and run the default workspace in minutes.
                    </Translate>
                </p>
                <div className="quickstart-blocks">
                    {quickStartBlocks.map((b) => (
                        <div key={b.id} className="quickstart-block">
                            <div className="quickstart-block-label">{b.label}</div>
                            <CodeBlock language="bash" className="quickstart-codeblock">
                                {b.body}
                            </CodeBlock>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

function FinalCta() {
    const githubImage = useBaseUrl('/img/github.svg');
    return (
        <section className="home-section home-section--cta">
            <div className="container text-center">
                <h2 className="section-title">
                    <Translate id="home.cta.title" description="Final CTA title">
                        Ready to ship AI-native software?
                    </Translate>
                </h2>
                <p className="section-lead">
                    <Translate id="home.cta.lead" description="Final CTA lead">
                        Explore the docs, review the source, or start building with LinaPro today.
                    </Translate>
                </p>
                <div className="cta-buttons">
                    <Link className="button button--primary button--md" to={DOCS_URL} style={{width: '200px'}}>
                        <Translate id="home.cta.docs" description="Final CTA docs button">
                            Read the Docs →
                        </Translate>
                    </Link>
                    <Link
                        className="button button--secondary button--md hover:bg-gray-200"
                        to={REPO_URL}
                        style={{
                            width: '200px',
                            paddingLeft: '50px',
                            backgroundImage: `url(${githubImage})`,
                            backgroundRepeat: 'no-repeat',
                            backgroundPosition: '22px center',
                        }}
                    >
                        <Translate id="home.cta.github" description="Final CTA github button">
                            Star on GitHub
                        </Translate>
                    </Link>
                </div>
            </div>
        </section>
    );
}

export default function Home(): JSX.Element {
    return (
        <Layout>
            <HomepageHeader />
            <Strengths />
            <Layers />
            <Workflow />
            <Modules />
            <QuickStart />
            <FinalCta />
        </Layout>
    );
}
