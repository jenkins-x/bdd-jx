import expect from 'expect';

describe('Smoke Tests', function() {
    context('project list', () => {
        let testProjectName = Cypress.env('APPLICATION_NAME')
        beforeEach(() => {
            cy.visit('/teams/jx/projects');
            cy.wait(6000)
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
        beforeEach( () => {
            cy.visit('/teams/jx/builds');
            cy.wait(6000)

            // Wait for build list to be loaded
            cy.get('[data-test=buildlist-build]').should('have.length.greaterThan', 0);
        });

        it('displays the details of the build', () => {
            // Get the build card by checking the build id
            cy.get('[data-test=buildlist-build-details]').and($div => {
               expect($div.text()).toContain('master');
               expect($div.text()).toContain('Unknown author');
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