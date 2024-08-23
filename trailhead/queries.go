package trailhead

// GetRankQuery returns GraphQL query for TrailheadRank
func GetRankQuery() string {
	return `
		fragment TrailheadRank on TrailheadRank {
			__typename
			title
			requiredPointsSum
			requiredBadgesCount
			imageUrl
		}

		fragment PublicProfile on PublicProfile {
			__typename
			trailheadStats {
				__typename
				earnedPointsSum
				earnedBadgesCount
				completedTrailCount
				rank {
                    ...TrailheadRank
                }
				nextRank {
                    ...TrailheadRank
                }
			}
		}

		query GetTrailheadRank($slug: String, $hasSlug: Boolean!) {
			profile(slug: $slug) @include(if: $hasSlug) {
				... on PublicProfile {
                    ...PublicProfile
                }
				... on PrivateProfile {
                    __typename
                }
			}
		}`
}

// GetSkillsQuery returns GraphQL query for EarnedSkill
func GetSkillsQuery() string {
	return `
        fragment EarnedSkill on EarnedSkill {
            __typename
            earnedPointsSum
            id
            itemProgressEntryCount
            skill {
                __typename
                apiName
                id
                name
            }
        }

        query GetEarnedSkills($slug: String, $hasSlug: Boolean!) {
            profile(slug: $slug) @include(if: $hasSlug) {
                __typename
                ... on PublicProfile {
                    id
                    earnedSkills {
                        ...EarnedSkill
                    }
                }
            }
        }`
}

// GetCertificationsQuery returns GraphQL query for GetUserCertifications
func GetCertificationsQuery() string {
	return `
        query GetUserCertifications($slug: String, $hasSlug: Boolean!) {
            profile(slug: $slug) @include(if: $hasSlug) {
                __typename
                id
                ... on PublicProfile {
                    credential {
                        messages {
                            __typename
                            body
                            header
                            location
                            image
                            cta {
                                __typename
                                label
                                url
                            }
                            orientation
                        }
                        messagesOnly
                        brands {
                            __typename
                            id
                            name
                            logo
                        }
                        certifications {
                            cta {
                                __typename
                                label
                                url
                            }
                            dateCompleted
                            dateExpired
                            downloadLogoUrl
                            logoUrl
                            infoUrl
                            maintenanceDueDate
                            product
                            publicDescription
                            status {
                                __typename
                                title
                                expired
                                date
                                color
                                order
                            }
                            title
                        }
                    }
                }
            }
        }`
}

// GetBadgesQuery returns GraphQL query for EarnedAward
func GetBadgesQuery() string {
	return `
        fragment EarnedAward on EarnedAwardBase {
            __typename
            id
            award {
                __typename
                id
                title
                type
                icon
                content {
                    __typename
                    webUrl
                    description
                }
            }
        }

        fragment EarnedAwardSelf on EarnedAwardSelf {
            __typename
            id
            award {
                __typename
                id
                title
                type
                icon
                content {
                    __typename
                    webUrl
                    description
                }
            }
            earnedAt
            earnedPointsSum
        }

        fragment StatsBadgeCount on TrailheadProfileStats {
            __typename
            earnedBadgesCount
            superbadgeCount
        }

        fragment ProfileBadges on PublicProfile {
            __typename
            trailheadStats {
                ... on TrailheadProfileStats {
                    ...StatsBadgeCount
                }
            }
            earnedAwards(first: $count, after: $after, awardType: $filter) {
                edges {
                    node {
                        ... on EarnedAwardBase {
                            ...EarnedAward
                        }
                        ... on EarnedAwardSelf {
                            ...EarnedAwardSelf
                        }
                    }
                }
                pageInfo {
                    ...PageInfoBidirectional
                }
            }
        }

        fragment PageInfoBidirectional on PageInfo {
            __typename
            endCursor
            hasNextPage
            startCursor
            hasPreviousPage
        }

        query GetTrailheadBadges(
            $slug: String
            $hasSlug: Boolean!
            $count: Int = 8
            $after: String = null
            $filter: AwardTypeFilter = null
        ) {
            profile(slug: $slug) @include(if: $hasSlug) {
                __typename
                ... on PublicProfile {
                    ...ProfileBadges
                }
            }
        }`
}
