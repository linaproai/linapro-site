import {test} from '../../fixtures/auth';
import {HomePage} from '../../pages/HomePage';

test.describe('TC-2 Chinese homepage workspace preview', () => {
  test('TC-2a: removes quickstart and applies dark display styling to workspace preview', async ({
    page,
  }) => {
    const homePage = new HomePage(page);

    await homePage.gotoZhHome();
    await homePage.expectZhHomeUsesDarkWorkspacePreviewWithoutQuickstart();
  });
});
