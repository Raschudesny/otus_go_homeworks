package hw04_lru_cache //nolint:golint,stylecheck

// List interface api.
type List interface {
	Len() int                          // returns current list len
	Front() *listItem                  // returns first elem from list
	Back() *listItem                   // returns last elem from list
	PushFront(v interface{}) *listItem // adds one element to list start
	PushBack(v interface{}) *listItem  // adds one element to list end
	Remove(i *listItem)                // removes list item from a list
	MoveToFront(i *listItem)           // moves list item to list start
}

type listItem struct {
	Value interface{}
	Next  *listItem
	Prev  *listItem
}

type list struct {
	// Place your code here
	size int
	head *listItem
	tail *listItem
}

func (l list) Len() int {
	return l.size
}

func (l list) Front() *listItem {
	return l.head
}

func (l list) Back() *listItem {
	return l.tail
}

func (l *list) PushFront(v interface{}) *listItem {
	if l.head == nil {
		l.head = &listItem{v, nil, nil}
		l.size++
		l.tail = l.head
	} else {
		prevHead := l.head
		l.head = &listItem{v, prevHead, nil}
		prevHead.Prev = l.head
		l.size++
	}
	return l.head
}

func (l *list) PushBack(v interface{}) *listItem {
	if l.tail == nil {
		l.tail = &listItem{v, nil, nil}
		l.size++
		l.head = l.tail
	} else {
		prevTail := l.tail
		l.tail = &listItem{v, nil, prevTail}
		prevTail.Next = l.tail
		l.size++
	}
	return l.tail
}

func (l *list) Remove(i *listItem) {
	prevItem := i.Prev
	nextItem := i.Next

	// meh... go-critic linter forces to use switch-case here
	switch {
	case prevItem == nil && nextItem == nil:
		l.head = nil
		l.tail = nil
	case prevItem == nil:
		l.head = nextItem
		nextItem.Prev = nil
	case nextItem == nil:
		l.tail = prevItem
		prevItem.Next = nil
	default:
		prevItem.Next = nextItem
		nextItem.Prev = prevItem
	}
	l.size--
}

func (l *list) MoveToFront(item *listItem) {
	l.Remove(item)
	if l.head == nil {
		l.head = item
		item.Next = nil
		item.Prev = nil
		l.tail = l.head
		l.size++
	} else {
		prevHead := l.head
		l.head = item
		item.Next = prevHead
		item.Prev = nil
		prevHead.Prev = l.head
		l.size++
	}
}

func NewList() List {
	return &list{}
}
