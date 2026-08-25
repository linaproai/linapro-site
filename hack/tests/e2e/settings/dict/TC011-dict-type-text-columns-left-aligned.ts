import { test, expect } from '../../../fixtures/auth';
import { DictPage } from '../../../pages/DictPage';

test.describe('TC-11 字典管理文本列左对齐', () => {
  test('TC-11a: 字典名称、字典类型和备注列居左展示', async ({
    adminPage,
  }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await expect(
      dictPage.typeHeader(/字典名称|Dictionary Name/i).first(),
    ).toBeVisible();
    expect(await dictPage.getTypeRowCount()).toBeGreaterThanOrEqual(1);

    const nameAlign = await dictPage.getColumnAlignment('type', '字典名称');
    expect(nameAlign.headerLeft).toBe(true);
    expect(nameAlign.bodyLeft).toBe(true);

    const typeAlign = await dictPage.getColumnAlignment('type', '字典类型');
    expect(typeAlign.headerLeft).toBe(true);
    expect(typeAlign.bodyLeft).toBe(true);

    const remarkAlign = await dictPage.getColumnAlignment('type', '备注');
    expect(remarkAlign.headerLeft).toBe(true);
    expect(remarkAlign.bodyLeft).toBe(true);

    const day = new Date().toISOString().slice(0, 10).replace(/-/g, '');
    const stamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
    await adminPage.screenshot({
      path: `../../temp/${day}/${stamp}-dict-type-text-columns-left.png`,
      fullPage: false,
    });
  });

  test('TC-11b: 字典数据列表备注列居左展示', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();
    await dictPage.clickTypeRow('sys_normal_disable');

    await expect(
      dictPage.dataHeader(/备注|Remark/i).first(),
    ).toBeVisible();
    expect(await dictPage.getDataRowCount()).toBeGreaterThanOrEqual(1);

    const remarkAlign = await dictPage.getColumnAlignment('data', '备注');
    expect(remarkAlign.headerLeft).toBe(true);
    expect(remarkAlign.bodyLeft).toBe(true);

    const day = new Date().toISOString().slice(0, 10).replace(/-/g, '');
    const stamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19);
    await adminPage.screenshot({
      path: `../../temp/${day}/${stamp}-dict-data-remark-column-left.png`,
      fullPage: false,
    });
  });
});
