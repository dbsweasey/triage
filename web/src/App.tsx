import { Layout } from './components/Layout'
import { ButtonLink, EmptyState } from './components/ui'
import { KnowledgeBase } from './pages/KnowledgeBase'
import { NewTicket } from './pages/NewTicket'
import { TicketDetail } from './pages/TicketDetail'
import { TicketList } from './pages/TicketList'
import { useRoute } from './router'

export default function App() {
  const route = useRoute()

  let page
  switch (route.name) {
    case 'tickets':
      page = <TicketList />
      break
    case 'new':
      page = <NewTicket />
      break
    case 'ticket':
      // key: remount when moving between tickets so state (like errors) doesn't carry over.
      page = <TicketDetail key={route.id} id={route.id} />
      break
    case 'kb':
      page = <KnowledgeBase />
      break
    default:
      page = (
        <EmptyState title="Page not found">
          <ButtonLink to="/">Back to tickets</ButtonLink>
        </EmptyState>
      )
  }

  return <Layout>{page}</Layout>
}
