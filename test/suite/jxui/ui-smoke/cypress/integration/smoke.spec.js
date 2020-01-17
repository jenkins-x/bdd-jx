import expect from 'expect';

describe('Smoke Tests', function() {
    context('project list', () => {
        let testProjectName = Cypress.env('APPLICATION_NAME')
        beforeEach(() => {
            cy.visit('/teams/jx/projects');

            // Wait for project list to be loaded
            cy.get('[data-test=projectlist-project]').should('have.length.greaterThan', 0);
        });

        it('displays the details of one project', () => {
            cy.get('[data-test=projectlist-project]').contains('[data-test=projectlist-project]', testProjectName).within(($div) => {
                cy.get('[data-test=projectlist-project-name]').and($div => {
                    expect($div.text()).toBe(testProjectName)
                });
                cy.get('[data-test=projectlist-project-details]').and($div => {
                    expect($div.text()).toContain(testProjectName)
                });
            })
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
        let testProjectName = Cypress.env('APPLICATION_NAME')
        beforeEach(() => {
            cy.visit('/teams/jx/builds');

            // Wait for build list to be loaded
            cy.get('[data-test=buildlist-build]').should('have.length.greaterThan', 0);
        });

        it('displays the details of the build', () => {
            // Get the build card by checking the build id
            cy.get('[data-test=buildlist-build-details]').contains('[data-test=buildlist-build-details]', testProjectName).parents('[data-test=buildlist-build]').within(() => {
                cy.get('[data-test=buildlist-build-details]').and($div => {
                    expect($div.text()).toContain('Author:');
                    expect($div.text()).toContain('Build started');
                    expect($div.text()).toContain('Success');
                })
            })
        });

        it('can search one build by name', () => {
            cy.get('[data-test=buildlist-search]').within(() => {
                cy.get('input').type(testProjectName);
            });
            cy.get('[data-test=buildlist-build-details]').should('have.length', 1);
            cy.get('[data-test=buildlist-build-details]').each(element => {
                expect(element.text()).toContain(testProjectName);
            });
        });
    })
})