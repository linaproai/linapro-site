import Giscus from "@giscus/react";
import { useEffect, useState } from "react";

type ColorMode = "light" | "dark";

function getHtmlColorMode(): ColorMode {
  if (typeof document === "undefined") {
    return "light";
  }

  return document.documentElement.getAttribute("data-theme") === "dark"
    ? "dark"
    : "light";
}

function useHtmlColorMode(): ColorMode {
  const [colorMode, setColorMode] = useState<ColorMode>("light");

  useEffect(() => {
    const html = document.documentElement;
    const syncColorMode = () => setColorMode(getHtmlColorMode());

    syncColorMode();

    const observer = new MutationObserver(syncColorMode);
    observer.observe(html, { attributes: true, attributeFilter: ["data-theme"] });

    return () => observer.disconnect();
  }, []);

  return colorMode;
}

// https://rikublock.dev/docs/tutorials/giscus-integration/
export default function Comments(): JSX.Element {
  const colorMode = useHtmlColorMode();

  return (
    <div className="docusaurus-mt-lg">
      <Giscus
        id="comments"
        repo="linaproai/linapro-site"
        repoId="R_kgDOSIbhHw"
        category="General"
        categoryId="DIC_kwDOSIbhH84C8cwX"
        mapping="pathname"
        strict="1"
        reactionsEnabled="1"
        emitMetadata="0"
        inputPosition="top"
        theme={colorMode === "dark" ? "dark_tritanopia" : "light_tritanopia"}
        lang="en"
        loading="lazy"
      />
    </div>
  );
}
