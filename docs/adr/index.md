# Architecture Decision Records

_ADR_ for short.

ADRs help document team decisions as a project evolves to provide context and insight into why significant decisions were made and help provide team members (new and existing) make new decisions.

ADRs document "Architecturally Significant Decisions".

## What is an Architecturally Significant Decision?

The idea is to capture key decisions having to do with anything _architectural_ in a way that promotes better communication than simple word-of-mouth.

What is an _architectural_ decision? Does the design decision:

- Alter externally visible system properties?
- Modify public interfaces?
- Directly influence high priority quality attributes?
- Include or remove dependencies?
- Result from a discussion where you learned more about technical or business constraints?
- Involve taking on strategic technical debt?
- Change the structures of the system (static, dynamic, or physical)?
- Require other developers to update construction techniques or development environments?

If yes to any of the above, it's probably architecturally significant.

Read [Documenting Architecture Decisions](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions) for more information.

## How to use ADRs

ADRs are meant to be a living document to record discussion, context, and decisions. The best practice for using ADRs is:

1. A person (or small group of people) create the ADR in a PR
1. The person(s) present the ADR to a group of stakeholders
1. The stakeholders discuss the ADR in a recorded manner (perhaps in the ADR's PR itself)
1. The person(s) updates the ADR to reflect the outcome of the group's discussion
1. Final review & merge the PR

If at any point the ADR is contradicted by a later group decision, it needs to be updated, or it becomes obsolete, it is the repository maintainer group's responsibility to revisit the ADR and present an adjustment to the group accordingly.

## How to create an ADR

The easiest way to get started is to look at recent ADRs for content and start from the [template](./000-template.md). The template is there to help you get started, not as a mandated format, feel free to adjust as needed. Create your ADR in `docs/adr` incrementing the number of the most recent ADR.

When you've written your ADR, create a [pull request](../../CONTRIBUTING.md). Project maintainers will review the ADR and when accepted, will be merged.

### Tips and Hints

- Titles should be descriptive, concise, and precise
- The whole document should be one or two pages long at most.
- Think of the document as a conversation with a future developer. This means write well and use full sentences. Avoid Rambling.
- Update consequences as they become known. The ADR becomes like a diary for seeing how the design decisions we make impact the system over time.
- Include diagrams as necessary. Many decisions don't require diagrams.

## References

Keeling, Michael; Runde, Joe. Architecture Decision Records in Action, from _Proceedings of SATURN2017 Conference_ [PDF](http://resources.sei.cmu.edu/library/asset-view.cfm?assetid=497744) [Video](https://www.youtube.com/watch?v=41NVge3_cYo)

Keeling, Michael; Runde, Joe. Share the Load: Distribute Design Authority with Architecture Decision Records, from _Agile 2018 in San Diego_ [Web](https://www.agilealliance.org/resources/experience-reports/distribute-design-authority-with-architecture-decision-records/) [Video](https://www.agilealliance.org/resources/sessions/share-the-load-distributing-design-authority-with-lightweight-decision-records/)

Nygard, Michael. Documenting Architecture Decisions, from _Think Relevance_ blog. [Web](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions)

Kruchten, Philippe. _The Decision View's Role in Software Architecture Practice_, IEEE Software 26:36-42, February 2009

Tyree, J. and Akerman, A. _Architecture Decisions: Demystifying Architecture_, IEEE Software 22:2:19-27, March-April 2005 [PDF](http://www.utdallas.edu/~chung/SA/zz-Impreso-architecture_decisions-tyree-05.pdf)
