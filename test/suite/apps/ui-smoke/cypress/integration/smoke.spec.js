import expect from 'expect';

describe('Smoke Tests', function() {
    context('project list', () => {
        let testProjectName = ''
        beforeEach(() => {
            cy.visit('/teams/jx/projects');

            cy.get('[data-test=projectlist-project-name]').first().and($div => {
                testProjectName = $div.text();
            });

            // Wait for testProjectName to be populated (means the page is loaded)
            cy.get('body').should(() => {
                expect(testProjectName).not.toBe('');
            });
        });

        it('contains projects', () => {
            cy.get('[data-test=projectlist-project]').should('have.length.greaterThan', 0);
        });

        it('displays the details of one project', () => {
            cy.get('[data-test=projectlist-project-name]').first().and($div => {
                expect($div.text()).not.toBe('')
                expect($div.text()).not.toBe(undefined);
            });
            cy.get('[data-test=projectlist-project-details]').first().and($div => {
                expect($div.text()).not.toBe('')
                expect($div.text()).not.toBe(undefined);
            });
        });

        it('can search one project by name', () => {            
            cy.get('[data-test=projectlist-search]').within(() => {
                cy.get('input').type(testProjectName);
            });
            cy.get('[data-test=projectlist-project]').should('have.length', 1);
            cy.get('[data-test=projectlist-project-name]').and($div => {
                expect($div.text()).toEqual(testProjectName);
            });
        });
    })

    context('build list', () => {
        let testProjectName = '';
        beforeEach(() => {
            cy.visit('/teams/jx/builds');

            cy.get('[data-test=buildlist-build-id]').first().and($div => {
                const buildId = $div.text();
                testProjectName = buildId.split('/')[1].replace(' ', '');
            });

            // Wait for testProjectName to be populated (means the page is loaded)
            cy.get('body').should(() => {
                expect(testProjectName).not.toBe('');
            });
        });

        it('contains builds', () => {
            cy.get('[data-test=buildlist-build]').should('have.length.greaterThan', 0);
        });

        it('displays the details of the build', () => {
            cy.get('[data-test=buildlist-build-id]').first().and($div => {
                expect($div.text()).not.toBe('')
                expect($div.text()).not.toBe(undefined);
            });
            cy.get('[data-test=buildlist-build-details]')
                .first()
                .and($div => {
                    expect($div.text()).toContain('Author:');
                    expect($div.text()).toContain('Build started');
                })
        });

        it('can search one build by name', () => {
            cy.get('[data-test=buildlist-search]').within(() => {
                cy.get('input').type(testProjectName);
            });
            cy.get('[data-test=buildlist-build-id]').each(element => {
                expect(element.text()).toContain(testProjectName);
            });
        });
    })
})